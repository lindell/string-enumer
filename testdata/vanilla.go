// extra-parameters: --type Test
package main

import (
	"encoding/json"
	"fmt"
)

// Test is a test type
type Test string

// Some Tests
const (
	TestTest  Test = "test"
	TestTest2 Test = "hello"
)

func main() {
	if ok := TestTest.ValidTest(); !ok {
		panic(fmt.Sprintf("should be valid Test"))
	}
	if ok := TestTest2.ValidTest(); !ok {
		panic(fmt.Sprintf("should be valid Test"))
	}

	nonValid := Test("test2")
	if ok := nonValid.ValidTest(); ok {
		panic(fmt.Sprintf("should not be valid Test"))
	}

	validRawJSON := []byte(`"hello"`)
	var test Test
	if err := json.Unmarshal(validRawJSON, &test); err != nil {
		panic(fmt.Sprintf("could not unmarshal: %s", err))
	}

	invalidRawJSON := []byte(`"test2"`)
	if err := json.Unmarshal(invalidRawJSON, &test); err != nil { // NB
		panic(fmt.Sprintf("could not unmarshal: %s", err))
	}
}
