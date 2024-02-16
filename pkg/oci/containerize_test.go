package oci

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func Test_validateLink(t *testing.T) {
	root := "testdata/test-links"

	tests := []struct {
		path  string // path of the file within test project root
		valid bool   // If it should be considered valid
		name  string // descriptive name of the test
	}{
		{"a.txt", true, "do not evaluate regular files"},
		{"a.lnk", true, "do not evaluate directories"},
		{"absoluteLink", false, "disallow absolute-path links"},
		{"a.lnk", true, "links to files within the root are allowed"},
		{"...validName.txt", true, "allow files with dot prefixes"},
		{"...validName.lnk", true, "allow links with target of dot prefixed names"},
		{"linkToRoot", true, "allow links to the project root"},
		{"b/linkToRoot", true, "allow links to the project root from within subdir"},
		{"b/linkToCurrentDir", true, "allow links to a subdirectory within the project"},
		{"b/linkToRootsParent", false, "disallow links to the project's immediate parent"},
		{"b/linkOutsideRootsParent", false, "disallow links outside project root and its parent"},
		{"b/c/linkToParent", true, "allow links up, but within project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(root, tt.path)
			info, err := os.Lstat(path) // filepath.Walk does not follow symlinks
			if err != nil {
				t.Fatal(err)
			}
			err = validateLink(root, path, info)
			if err == nil != tt.valid {
				t.Fatalf("expected %v, got %v", tt.valid, err)
			}
		})
	}
	// Run a windows-specific absolute path test
	// Note this technically succeeds on unix systems, but wrapping in
	// an runtime check seems like a good idea to make it more clear.
	if runtime.GOOS != "windows" {
		path := "c://some/absolute/path"
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		err = validateLink(root, path, info)
		if err == nil {
			t.Fatal("absolute path should be invalid on windows")
		}
	}
}
