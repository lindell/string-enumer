// Package stringenumer generate useful code for enums defined as strings
//
//	type MyEnum string
//	const (
//		MyEnumThis MyEnum = "this"
//		MyEnumThat MyEnum = "that"
//	)
//
// The code that is generated by default is a ValidMyEnum function:
//	ValidMyEnum(string) bool
// It can also generate an TextUnmarshaling function for the type that validates any string that is unmarshaled into this type.
// Via, for example, json.Unmarshal.
package stringenumer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"log"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/tools/go/packages"
)

// Option is used in the call of Generate to change the behavior
type Option func(*generator)

// TypeNames is the options used to set the types that code should be generated for
// At least one type needs to be defined
func TypeNames(types ...string) Option {
	return func(g *generator) {
		g.typeNames = append(g.typeNames, types...)
	}
}

// TextUnmarshaling sets if text unarshaling should be generated or not
func TextUnmarshaling(unmarshalText bool) Option {
	return func(g *generator) {
		g.unmarshalText = unmarshalText
	}
}

// Paths sets the paths from where code should be read from
func Paths(paths ...string) Option {
	return func(g *generator) {
		g.paths = paths
	}
}

type multiError []error

func (m multiError) Error() string {
	errors := make([]string, len(m))
	for i := range m {
		errors[i] = m[i].Error()
	}
	return strings.Join(errors, "; ")
}

// Generate returns a reader with generated code
func Generate(options ...Option) (io.Reader, error) {
	g := generator{
		values:  map[string][]value{},
		imports: map[string]struct{}{},
	}

	for _, option := range options {
		option(&g)
	}

	g.parsePackage(g.paths...)
	g.parseFiles()

	if len(g.errors) != 0 {
		return nil, g.errors
	}

	if err := g.validateValues(); err != nil {
		return nil, err
	}

	for _, typename := range g.typenames() {
		g.buildBasics(typename)
		if g.unmarshalText {
			g.buildTextUnmarshaling(typename)
		}
	}

	g.buildHeader()

	return io.MultiReader(&g.headerBuf, &g.buf), nil
}

// file holds a single parsed file and associated data.
type file struct {
	pkg  *pkg      // Package to which this file belongs.
	file *ast.File // Parsed AST.
}

// value represents a declared constant.
type value struct {
	name  string
	value string
}

// pkg holds information about a Go package
type pkg struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*file
}

type generator struct {
	typeNames []string
	paths     []string

	buf    bytes.Buffer
	pkg    *pkg       // Package we are scanning.
	errors multiError // Errors while parsing or processing the file
	// Accumulator for constant values of that type. The key is the name, and the value is all values of that type
	values map[string][]value

	unmarshalText bool

	imports   map[string]struct{}
	headerBuf bytes.Buffer
}

// Printf prints the string to the output
func (g *generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *generator) parsePackage(patterns ...string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

// typenames return a list of all type names
func (g *generator) typenames() []string {
	ret := make([]string, 0, len(g.values))
	for v := range g.values {
		ret = append(ret, v)
	}
	sort.Strings(ret)
	return ret
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *generator) addPackage(p *packages.Package) {
	g.pkg = &pkg{
		name:  p.Name,
		defs:  p.TypesInfo.Defs,
		files: make([]*file, len(p.Syntax)),
	}

	for i, f := range p.Syntax {
		g.pkg.files[i] = &file{
			file: f,
			pkg:  g.pkg,
		}
	}
}

func (g *generator) isTypeName(tn string) bool {
	for _, typeName := range g.typeNames {
		if typeName == tn {
			return true
		}
	}
	return false
}

func (g *generator) parseFiles() {
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		if file.file != nil {
			ast.Inspect(file.file, g.genDecl(file))
		}
	}
}

func (g *generator) addImport(s string) {
	g.imports[s] = struct{}{}
}

// genDecl processes one declaration clause.
func (g *generator) genDecl(f *file) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		decl, ok := node.(*ast.GenDecl)
		if !ok || decl.Tok != token.CONST {
			// We only care about const declarations.
			return true
		}
		// The name of the type of the constants we are declaring.
		// Can change if this is a multi-element declaration.
		typ := ""
		// Loop over the elements of the declaration. Each element is a ValueSpec:
		// a list of names possibly followed by a type, possibly followed by values.
		// If the type and value are both missing, we carry down the type (and value,
		// but the "go/types" package takes care of that).
		for _, spec := range decl.Specs {
			vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
			if vspec.Type == nil && len(vspec.Values) > 0 {
				// "X = 1". With no type but a value, the constant is untyped.
				// Skip this vspec and reset the remembered type.
				typ = ""
				continue
			}
			if vspec.Type != nil {
				// "X T". We have a type. Remember it.
				ident, ok := vspec.Type.(*ast.Ident)
				if !ok {
					continue
				}
				typ = ident.Name
			}
			if !g.isTypeName(typ) {
				// This is not the type we're looking for.
				continue
			}
			// We now have a list of names (from one line of source code) all being
			// declared with the desired type.
			// Grab their names and actual values and store them in g.values.
			for _, name := range vspec.Names {
				if name.Name == "_" {
					continue
				}
				// This dance lets the type checker find the values for us. It's a
				// bit tricky: look up the object declared by the name, find its
				// types.Const, and extract its value.
				obj, ok := f.pkg.defs[name]
				if !ok {
					g.errors = append(g.errors, fmt.Errorf("no value for constant %s", name))
					return false
				}
				info := obj.Type().Underlying().(*types.Basic).Info()
				if info&types.IsString == 0 {
					g.errors = append(g.errors, fmt.Errorf("can't handle non-string constant type %s", typ))
					return false
				}
				val := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
				if val.Kind() != constant.String {
					g.errors = append(g.errors, fmt.Errorf("can't happen: constant is not an string %s", name))
					return false
				}
				str := constant.StringVal(val)

				v := value{
					name:  name.Name,
					value: str,
				}
				g.values[typ] = append(g.values[typ], v)
			}
		}
		return false
	}
}

// validateValues ensures that there exist no more than one value of each type
func (g *generator) validateValues() error {
	var errors multiError
	for typeName, v := range g.values {
		values := map[string]struct{}{}
		for _, value := range v {
			if _, ok := values[value.value]; ok {
				errors = append(errors, fmt.Errorf("the type %s has multiple values of %s", typeName, value.value))
				break
			}
			values[value.value] = struct{}{}
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

func (g *generator) buildHeader() {
	fmt.Fprintf(&g.headerBuf, "package %s\n", g.pkg.name)

	if len(g.imports) > 0 {
		fmt.Fprintln(&g.headerBuf, "\nimport (")

		// Sort imports by name
		imports := make([]string, 0, len(g.imports))
		for imp := range g.imports {
			imports = append(imports, imp)
		}

		for _, imp := range imports {
			fmt.Fprintln(&g.headerBuf, "	"+imp)
		}
		fmt.Fprint(&g.headerBuf, ")\n")
	}
}

func (g *generator) buildBasics(name string) {
	values := g.values[name]
	g.Printf("\n// valid%sValues contains a map of all valid %s values for easy lookup\n", strings.Title(name), name)
	g.Printf("var valid%sValues = map[%s]struct{}{\n", strings.Title(name), name)
	maxNameLength := maxNameLength(values)
	for _, v := range values {
		g.Printf("	%s: %sstruct{}{},\n", v.name, strings.Repeat(" ", maxNameLength-utf8.RuneCountInString(v.name)))
	}
	g.Printf("}\n\n")
	g.Printf("// Valid%s validates if a value is a valid %s\n", strings.Title(name), name)
	g.Printf("func (v %s) Valid%s() bool {\n", name, strings.Title(name))
	g.Printf("	_, ok := valid%sValues[v]\n", strings.Title(name))
	g.Printf("	return ok\n")
	g.Printf("}\n\n")
	g.Printf("// %sValues returns a list of all (valid) %s values\n", strings.Title(name), name)
	g.Printf("func %sValues() []%s {\n", strings.Title(name), name)
	g.Printf("	return []%s{\n", name)
	for _, v := range values {
		g.Printf("		%s,\n", v.name)
	}
	g.Printf("	}\n")
	g.Printf("}\n")
}

func (g *generator) buildTextUnmarshaling(name string) {
	g.addImport(`"fmt"`)
	g.Printf("\n// UnmarshalText takes a text, verifies that it is a correct %s and unmarshals it\n", name)
	g.Printf("func (v *%s) UnmarshalText(text []byte) error {\n", strings.Title(name))
	g.Printf("	if valid := %s(text).Valid%s(); !valid {\n", strings.Title(name), name)
	g.Printf("		return fmt.Errorf(\"not valid value for %s: %%s\", text)\n", name)
	g.Printf("	}\n")
	g.Printf("	*v = %s(text)\n", name)
	g.Printf("	return nil\n")
	g.Printf("}\n")
}

func maxNameLength(vv []value) int {
	max := 0
	for _, v := range vv {
		if len(v.name) > max {
			max = utf8.RuneCountInString(v.name)
		}
	}
	return max
}
