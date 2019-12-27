// extra-parameters: --type Test --type Test2
package main

import (
	"fmt"
)

// Test is a test type
type Test string

// Some Tests
const (
	TestTest  Test = "test"
	TestTest2 Test = "hello"
)

// Test2 is a test type
type Test2 string

// Some Test2s
const (
	Test2Test  Test2 = "test"
	Test2Test2 Test2 = "hello"
	Test2Test3 Test2 = "test3"
)

func main() {
	if ok := TestTest.ValidTest(); !ok {
		panic(fmt.Sprintf("should be valid Test"))
	}
	if ok := TestTest2.ValidTest(); !ok {
		panic(fmt.Sprintf("should be valid Test"))
	}
	if ok := Test2Test.ValidTest2(); !ok {
		panic(fmt.Sprintf("should be valid Test2"))
	}
	if ok := Test2Test2.ValidTest2(); !ok {
		panic(fmt.Sprintf("should be valid Test2"))
	}
	if ok := Test2Test3.ValidTest2(); !ok {
		panic(fmt.Sprintf("should be valid Test2"))
	}

	nonValid := Test("test3")
	if ok := nonValid.ValidTest(); ok {
		panic(fmt.Sprintf("should not be valid Test"))
	}

	valid := Test2("test3")
	if ok := valid.ValidTest2(); !ok {
		panic(fmt.Sprintf("should be valid Test2"))
	}
}
