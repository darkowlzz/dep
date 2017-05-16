// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/dep"
	"github.com/golang/dep/internal/test"
)

const testGodepProjectRoot = "github.com/golang/notexist"

func TestGodepJsonLoad(t *testing.T) {
	// This is same as cmd/dep/testdata/Godeps.json
	wantJSON := godepJson{
		Name: "github.com/golang/notexist",
		Imports: []godepPackage{
			{
				ImportPath: "github.com/sdboyer/deptest",
				Rev:        "3f4c3bea144e112a69bbe5d8d01c1b09a544253f",
			},
			{
				ImportPath: "github.com/sdboyer/deptestdos",
				Rev:        "5c607206be5decd28e6263ffffdcee067266015e",
				Comment:    "v2.0.0",
			},
		},
	}

	h := test.NewHelper(t)
	defer h.Cleanup()

	h.TempCopy(filepath.Join(testGodepProjectRoot, "Godeps", godepJsonName), "Godeps.json")

	projectRoot := h.Path(testGodepProjectRoot)

	g := &godepFile{
		loggers: &dep.Loggers{
			Out:     log.New(os.Stderr, "", 0),
			Err:     log.New(os.Stderr, "", 0),
			Verbose: true,
		},
	}
	err := g.load(projectRoot)
	if err != nil {
		t.Fatalf("Error while loading... %v", err)
	}

	if g.json.Name != wantJSON.Name {
		t.Fatalf("Expected project name to be %v, but got %v", wantJSON.Name, g.json.Name)
	}

	if !equalImports(g.json.Imports, wantJSON.Imports) {
		t.Fatalf("Expected imports to be equal. \n\t(GOT): %v\n\t(WNT): %v", g.json.Imports, wantJSON.Imports)
	}
}

func TestGodepConvertProject(t *testing.T) {
	loggers := &dep.Loggers{
		Out:     log.New(os.Stdout, "", 0),
		Err:     log.New(os.Stderr, "", 0),
		Verbose: true,
	}

	f := godepFile{
		loggers: loggers,
		json: godepJson{
			Name: "github.com/foo/bar",
			Imports: []godepPackage{
				{
					ImportPath: "github.com/sdboyer/deptest",
					Rev:        "6a741be0cc55ecbe4f45690ebfd606a956d5f14a",
					Comment:    "v1.0.0",
				},
			},
		},
	}

	manifest, lock, err := f.convert("", nil)
	if err != nil {
		t.Fatal(err)
	}

	d, ok := manifest.Dependencies["github.com/sdboyer/deptest"]
	if !ok {
		t.Fatal("Expected the manifest to have a dependency for 'github.com/sdboyer/deptest' but got none")
	}

	v := d.Constraint.String()
	if v != "v1.0.0" {
		t.Fatalf("Expected manifest constraint to be master, got %s", v)
	}

	p := lock.P[0]
	if p.Ident().ProjectRoot != "github.com/sdboyer/deptest" {
		t.Fatalf("Expected the lock to have a project for 'github.com/sdboyer/deptest' but got '%s'", p.Ident().ProjectRoot)
	}

	lv := p.Version().String()
	if lv != "v1.0.0" {
		t.Fatalf("Expected locked revision to be 'v1.0.0', got %s", lv)
	}
}

func TestGodepConvertBadInput_EmptyPackageName(t *testing.T) {
	loggers := &dep.Loggers{
		Out:     log.New(os.Stdout, "", 0),
		Err:     log.New(os.Stderr, "", 0),
		Verbose: true,
	}

	f := godepFile{
		loggers: loggers,
		json: godepJson{
			Imports: []godepPackage{{ImportPath: ""}},
		},
	}

	_, _, err := f.convert("", nil)
	if err == nil {
		t.Fatal("Expected conversion to fail because the ImportPath is empty")
	}
}

// Compares two slices of godepPackage and checks if they are equal.
func equalImports(a, b []godepPackage) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
