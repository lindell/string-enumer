<h1 align="center">
  🧵 string-enumer
</h1>

<div align="center">
	<a href="https://github.com/lindell/string-enumer/actions?query=branch%3Amaster+workflow%3AGo"><img alt="Go build status" src="https://github.com/lindell/string-enumer/workflows/Go/badge.svg?branch=master" /></a>
	<a href="https://godoc.org/github.com/lindell/string-enumer/pkg/stringenumer"><img alt="GoDoc" src="https://godoc.org/github.com/lindell/string-enumer/pkg/stringenumer?status.svg" /></a>
	<a href="https://goreportcard.com/report/github.com/lindell/string-enumer"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/lindell/string-enumer" /></a>
</div>
<br>

String enumer is a golang code generator for enums declared as strings.

The function `func (v X) ValidX() bool` will always be generated on the defined types together with `func XValues() []X`. But options to generate more code exist.
It is especially useful with the `--text` option, that generates an `UnmarshalText` function which forces any unmarshaling of the type (via for JSON/XML/etc.) to be limited to the defined types.

The tool is primarily intended to be used with [go:generate](https://blog.golang.org/generate), but can be used as a separate CLI tool.

# Example usage with go generate

```go
//go:generate string-enumer --text -t Country -o ./generated.go .
// or
//go:generate go run github.com/lindell/string-enumer --text -t Country -o ./generated.go .
type Country string

const (
	CountryCanada       Country = "CA"
	CountryChina        Country = "CN"
	CountrySweden       Country = "SE"
	CountryUnitedStates Country = "US"
)
```

Will generate:

```go
// ValidCountry validates if a value is a valid Country
func (v Country) ValidCountry() bool {
	...
}

// CountryValues returns a list of all (valid) Country values
func CountryValues() []Country {
	...
}

// UnmarshalText takes a text, verifies that it is a correct Country and unmarshals it
func (v *Country) UnmarshalText(text []byte) error {
	...
}
```

## CLI Description:

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
