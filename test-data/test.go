package test

type Test string

const (
	TestTest  Test = "test"
	TestHello Test = "hello"
)

type Test2 string

const (
	TestTest2  Test2 = "test"
	TestHello2 Test  = "hello"
	TestTest3  Test2 = "test"
)
