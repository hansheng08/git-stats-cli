package archtest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// findRepoRoot walks up from the current directory until it finds go.mod.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// TestArchitecture checks the whole repo against ARCHITECTURE §5.1 and §11.
func TestArchitecture(t *testing.T) {
	root := findRepoRoot(t)
	goMod := filepath.Join(root, "go.mod")
	modulePath, err := readModulePath(goMod)
	if err != nil {
		t.Fatal(err)
	}
	got, err := check(root, modulePath, goMod)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range got {
		t.Error(fmt.Sprint(v))
	}
}
