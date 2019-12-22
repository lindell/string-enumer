string-enumer
----

String enumer is a code generator for enums declared as strings, like the example below:

```go
type MyType string

const (
    MyTypeThis MyType = "this"
    MyTypeThat MyType = "that"
)
```

The function `ValidMyType(string) bool` will always be generated. But options to generate more code exist.

The tool is primarily intended to be used with `go:generate`, but can be used as a package together with other go code.

# Example usage with go generate

```go
//go:generate string-enumer --text -t MyType -t YourType -o ./generated.go .
// or 
//go:generate github.com/lindell/string-enumer --text -t MyType -t YourType -o ./generated.go .
type MyType string

const (
    MyTypeThis MyType = "this"
    MyTypeThat MyType = "that"
)

type YourType string

const (
    YourTypeThis YourType = "this"
    YourTypeThat YourType = "that"
)
```

CLI Description:
----
```
$ string-enumer --help
Usage of string-enumer:
	string-enumer [flags] --type T --type T2 [directory]
	string-enumer [flags] --type T --type T2 files... # Must be a single package
For more information, see:
	https://github.com/lindell/string-enumer
Flags:
  -o, --output string   output file name; default is stdout
  -T, --text            if set, text unmarshaling methods will be generated. Default: false
  -t, --type strings    the type name(s), can be multiple, but at least on must be set
```
