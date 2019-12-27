package stringenumer

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestPurity(t *testing.T) {
	r, err := Generate(
		Paths("../../testdata/multiple.go"),
		TypeNames("Test", "Test2"),
		TextUnmarshaling(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	initialData, _ := ioutil.ReadAll(r)
	for i := 0; i < 20; i++ {
		r, err := Generate(
			Paths("../../testdata/multiple.go"),
			TypeNames("Test", "Test2"),
			TextUnmarshaling(true),
		)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := ioutil.ReadAll(r)

		if !bytes.Equal(data, initialData) {
			t.Fatal("different runs of Generate does not result in the same output")
		}
	}
}
