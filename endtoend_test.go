package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "string-enumer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "string-enumer")

	// Compile
	run("go", "build", "-o", binPath)
	if err != nil {
		t.Fatalf("building string-enumer: %s", err)
	}

	names, err := readDir("testdata")
	if err != nil {
		t.Fatalf("could not read files in directory: %s", err)
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			compileAndRun(t, tmpDir, binPath, name)
		})
	}
}

func compileAndRun(t *testing.T, dir, binPath, fileName string) {
	t.Logf("run: %s\n", fileName)

	sourcePath := filepath.Join(dir, fileName)

	// Copy file to the temporary directory
	err := copy(sourcePath, filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatalf("copying file to temporary directory: %s", err)
	}

	// Get parameters (except input and output file to be used)
	extraParameters, err := getExtraParameters(sourcePath)
	if err != nil {
		t.Fatalf("copying file to temporary directory: %s", err)
	}

	outputName := fmt.Sprintf("%d", rand.Int())
	outputPath := filepath.Join(dir, outputName+"_output.go")

	// Run the code generation
	params := []string{"--output", outputPath, sourcePath}
	params = append(params, extraParameters...)
	err = run(binPath, params...)
	if err != nil {
		t.Fatalf("could not run string-enumer: %s", err)
	}

	// Run the main() function in the source file, with the generated code attached
	err = run("go", "run", sourcePath, outputPath)
	if err != nil {
		t.Fatal(err)
	}
}

var extraParameterRegexp = regexp.MustCompile("// extra-parameters: ([^\n]+)")

// getExtraParameters gets extra parameters specified in a source file
func getExtraParameters(filepath string) ([]string, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	match := extraParameterRegexp.FindSubmatch(b)
	if match == nil {
		return nil, nil
	}

	return strings.Split(string(match[1]), " "), nil
}

// readDir reads and returns all files in a directory
func readDir(path string) ([]string, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	return fd.Readdirnames(-1)
}

// copy copies the from file to the to file.
func copy(to, from string) error {
	toFd, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFd.Close()
	fromFd, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFd.Close()
	_, err = io.Copy(toFd, fromFd)
	return err
}

// run runs a single command and returns an error if it does not succeed.
// os/exec should have this function, to be honest.
func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
