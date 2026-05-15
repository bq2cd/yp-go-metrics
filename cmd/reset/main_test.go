package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

const (
	notGeneratedSuffix = ":not-generated"
)

func TestRun(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("cannot resolve testdata path: %v", err)
	}

	tmpdir := t.TempDir()

	initGoModule(t, tmpdir, "testdata")

	err = os.CopyFS(tmpdir, os.DirFS(filepath.Join(testdata, "src")))
	if err != nil {
		t.Fatalf("cannot copy testdata/src to temp dir: %v", err)
	}

	// Act
	err = run(t.Context(), Config{StartDir: tmpdir})
	if err != nil {
		t.Fatalf("code generation failed: %v", err)
	}

	// Assert
	v := &validator{
		T:         t,
		SrcDir:    tmpdir,
		GoldenDir: filepath.Join(testdata, "golden"),
	}

	v.ValidateGeneratedFiles()
}

func initGoModule(t *testing.T, dir, name string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "go", "mod", "init", name)
	cmd.Dir = dir

	err := cmd.Run()
	if err != nil {
		t.Fatalf("cannot init go module at %s: %v", dir, err)
	}
}

type validator struct {
	T         *testing.T
	SrcDir    string
	GoldenDir string
}

func (v *validator) ValidateGeneratedFiles() {
	v.T.Helper()

	err := fs.WalkDir(os.DirFS(v.GoldenDir), ".", v.walk)
	if err != nil {
		v.T.Fatalf("cannot walk golden dir %s: %v", v.GoldenDir, err)
	}
}

func (v *validator) walk(path string, entry fs.DirEntry, werr error) error {
	if werr != nil {
		return werr
	}

	if entry.IsDir() {
		return nil // keep walking
	}

	switch entry.Name() {
	case outputFilename:
		v.assertEqualContent(path)
	case outputFilename + notGeneratedSuffix:
		v.assertNotGenerated(strings.TrimSuffix(path, notGeneratedSuffix))
	}

	return nil
}

func (v *validator) assertEqualContent(relpath string) {
	v.T.Helper()

	v.T.Logf("validating content of %s", relpath)

	goldenPath := filepath.Join(v.GoldenDir, relpath)
	generatedPath := filepath.Join(v.SrcDir, relpath)

	goldenContent, err := os.ReadFile(goldenPath)
	if err != nil {
		v.T.Errorf("cannot read golden file at %s: %v", goldenPath, err)

		return
	}

	generatedContent, err := os.ReadFile(generatedPath)
	if err != nil {
		v.T.Errorf("cannot read generated file at %s: %v", generatedPath, err)

		return
	}

	diff := gocmp.Diff(
		bytes.TrimSpace(generatedContent),
		bytes.TrimSpace(goldenContent),
	)
	if diff != "" {
		v.T.Errorf("generated file differs from golden file: %v", diff)
	}
}

func (v *validator) assertNotGenerated(relpath string) {
	v.T.Helper()

	v.T.Logf("ensuring %s is missing", relpath)

	abspath := filepath.Join(v.SrcDir, relpath)

	_, err := os.Lstat(abspath)
	if !os.IsNotExist(err) {
		v.T.Errorf("expected %s to be missing", relpath)
	}
}
