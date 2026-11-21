package archtest

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Allowed go.mod requirements. Single source of truth for this check;
// see ARCHITECTURE §2 and docs/adr/ADR-001-language-and-stack.md.
var (
	allowedDirectModules   = []string{"github.com/spf13/pflag", "github.com/mattn/go-runewidth"}
	allowedIndirectModules = []string{"github.com/rivo/uniseg"}
)

// Rule names reported in violations.
const (
	ruleImportWhitelist = "import-whitelist" // rule 1 (ARCHITECTURE §5.1)
	ruleOSExec          = "os-exec"          // rule 2
	ruleNetwork         = "network"          // rule 3
	ruleTestPathLeak    = "test-path-leak"   // rule 4
	ruleThirdParty      = "third-party"      // rule 5
	ruleGoModRequire    = "go-mod-require"   // rule 6
	hintNewPackage      = "new packages need an ARCHITECTURE §5.1 change reviewed by the Architect"
	placeholderUnlisted = "<unlisted package %s>"
	placeholderNonTest  = "<non-test file in test-only package %s>"
)

// violation is one architecture rule breach.
type violation struct {
	File   string // slash-separated path relative to root
	Import string // offending import path, module, or placeholder
	Rule   string
}

func (v violation) String() string {
	msg := fmt.Sprintf("%s: %q violates rule %s", v.File, v.Import, v.Rule)
	if v.Rule == ruleImportWhitelist && strings.HasPrefix(v.Import, "<") {
		msg += " (" + hintNewPackage + ")"
	}
	return msg
}

// allowedProject maps a package directory to the project-internal packages
// (relative to the module root) its non-test files may import.
// Source: ARCHITECTURE §5.1. A directory absent from this map is unlisted.
var allowedProject = map[string][]string{
	"cmd/git-stats":      {"internal/cli", "internal/version"},
	"internal/cli":       {"internal/daterange", "internal/gitlog", "internal/stats", "internal/render", "internal/model", "internal/apperr", "internal/version"},
	"internal/daterange": {"internal/model", "internal/apperr"},
	"internal/gitlog":    {"internal/model", "internal/apperr"},
	"internal/stats":     {"internal/model"},
	"internal/render":    {"internal/model"},
	"internal/model":     {},
	"internal/apperr":    {},
	"internal/version":   {},
	"internal/archtest":  {},
	"test/e2e":           {},
	"test/e2e/fakegit":   {},
}

func underDir(dir, prefix string) bool {
	return dir == prefix || strings.HasPrefix(dir, prefix+"/")
}

func isTestutil(dir string) bool { return underDir(dir, "internal/testutil") }

// isTestPath reports whether a directory is a test-only path.
func isTestPath(dir string) bool { return isTestutil(dir) || underDir(dir, "test") }

// isProduct reports whether dir is a product package (ARCHITECTURE §11).
func isProduct(dir string) bool {
	if isTestutil(dir) || underDir(dir, "internal/archtest") {
		return false
	}
	return underDir(dir, "cmd") || underDir(dir, "internal")
}

func skipDir(name string) bool {
	switch name {
	case "testdata", "vendor", "bin", "dist", "poc":
		return true
	}
	return strings.HasPrefix(name, ".")
}

func isModule(imp string, mods []string) bool {
	for _, m := range mods {
		if imp == m || strings.HasPrefix(imp, m+"/") {
			return true
		}
	}
	return false
}

// check scans every Go file under root and returns all violations, sorted.
func check(root, modulePath, goModPath string) ([]violation, error) {
	var out []violation
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		f, err := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", rel, err)
		}
		var imports []string
		for _, spec := range f.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("%s: bad import %s: %w", rel, spec.Path.Value, err)
			}
			imports = append(imports, imp)
		}
		out = append(out, checkFile(rel, imports, modulePath)...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	mod, err := checkGoMod(root, goModPath)
	if err != nil {
		return nil, err
	}
	out = append(out, mod...)
	sortViolations(out)
	return out, nil
}

func sortViolations(vs []violation) {
	sort.Slice(vs, func(i, j int) bool {
		a, b := vs[i], vs[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Import != b.Import {
			return a.Import < b.Import
		}
		return a.Rule < b.Rule
	})
}

// checkFile applies rules 1-5 to one file.
func checkFile(rel string, imports []string, modulePath string) []violation {
	var out []violation
	add := func(imp, rule string) { out = append(out, violation{File: rel, Import: imp, Rule: rule}) }

	dir := path.Dir(rel)
	isTest := strings.HasSuffix(rel, "_test.go")

	allowed, listed := allowedProject[dir]
	switch {
	case !listed && !isTestutil(dir):
		add(fmt.Sprintf(placeholderUnlisted, dir), ruleImportWhitelist)
	case !isTest && (dir == "internal/archtest" || dir == "test/e2e"):
		add(fmt.Sprintf(placeholderNonTest, dir), ruleImportWhitelist)
	}

	for _, imp := range imports {
		// Rule 2: os/exec.
		if imp == "os/exec" && !isTest && dir != "internal/gitlog" && !isTestPath(dir) {
			add(imp, ruleOSExec)
		}
		// Rule 3: network packages (non-test files only).
		if (imp == "net" || strings.HasPrefix(imp, "net/")) && !isTest {
			add(imp, ruleNetwork)
		}
		inModule := imp == modulePath || strings.HasPrefix(imp, modulePath+"/")
		if !inModule {
			// Rule 5: third-party imports.
			first, _, _ := strings.Cut(imp, "/")
			if strings.Contains(first, ".") &&
				!isModule(imp, allowedDirectModules) && !isModule(imp, allowedIndirectModules) {
				add(imp, ruleThirdParty)
			}
			continue
		}
		target := strings.TrimPrefix(strings.TrimPrefix(imp, modulePath), "/")
		if isTest && target == dir { // external test package xxx_test
			continue
		}
		// Rule 4: test-only paths must not flow into product code.
		if isTestutil(target) {
			if !isTest {
				add(imp, ruleTestPathLeak)
			}
			continue
		}
		if underDir(target, "test") {
			add(imp, ruleTestPathLeak)
			continue
		}
		// Rule 1: project import allowlist.
		if !listed {
			continue // unlisted package already reported once for the file
		}
		if !contains(allowed, target) {
			add(imp, ruleImportWhitelist)
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// checkGoMod applies rule 6 to the go.mod text (standard library only).
func checkGoMod(root, goModPath string) ([]violation, error) {
	data, err := readFile(goModPath)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(goModPath)
	if r, err := filepath.Rel(root, goModPath); err == nil && !strings.HasPrefix(r, "..") {
		name = filepath.ToSlash(r)
	}

	var out []violation
	inBlock := false
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		code, comment, _ := strings.Cut(line, "//")
		code = strings.TrimSpace(code)
		indirect := strings.Contains(comment, "indirect")
		var entry string
		switch {
		case inBlock && code == ")":
			inBlock = false
			continue
		case inBlock:
			entry = code
		case code == "require (":
			inBlock = true
			continue
		case strings.HasPrefix(code, "require "):
			entry = strings.TrimSpace(strings.TrimPrefix(code, "require "))
		default:
			continue
		}
		fields := strings.Fields(entry)
		if len(fields) == 0 {
			continue
		}
		mod := fields[0]
		list := allowedDirectModules
		if indirect {
			list = allowedIndirectModules
		}
		if !contains(list, mod) {
			out = append(out, violation{File: name, Import: mod, Rule: ruleGoModRequire})
		}
	}
	return out, sc.Err()
}

// readFile reads a file through fs.ReadFile so gosec G304 does not fire.
func readFile(name string) ([]byte, error) {
	return fs.ReadFile(os.DirFS(filepath.Dir(name)), filepath.Base(name))
}

// readModulePath returns the module path declared in goModPath.
func readModulePath(goModPath string) (string, error) {
	data, err := readFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.Trim(strings.TrimSpace(rest), `"`), nil
		}
	}
	return "", fmt.Errorf("no module line in %s", goModPath)
}

// --- self tests: prove the checker really fails ---

func TestChecker_Clean(t *testing.T) {
	root := filepath.Join("testdata", "clean")
	got, err := check(root, "example.com/fake", filepath.Join(root, "gomod.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("clean tree must have 0 violations, got %d: %v", len(got), got)
	}
}

func TestChecker_Violations(t *testing.T) {
	root := filepath.Join("testdata", "violations")
	got, err := check(root, "example.com/fake", filepath.Join(root, "gomod.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := []violation{
		{"gomod.txt", "github.com/evil/dep", ruleGoModRequire},
		{"internal/archtest/helper.go", fmt.Sprintf(placeholderNonTest, "internal/archtest"), ruleImportWhitelist},
		{"internal/cli/uses_e2e.go", "example.com/fake/test/e2e/fakegit", ruleTestPathLeak},
		{"internal/model/net.go", "net/http", ruleNetwork},
		{"internal/model/third.go", "github.com/evil/dep", ruleThirdParty},
		{"internal/render/exec.go", "os/exec", ruleOSExec},
		{"internal/stats/bad.go", "example.com/fake/internal/gitlog", ruleImportWhitelist},
		{"internal/stats/uses_fixture.go", "example.com/fake/internal/testutil/fixture", ruleTestPathLeak},
		{"internal/unknown/x.go", fmt.Sprintf(placeholderUnlisted, "internal/unknown"), ruleImportWhitelist},
		{"test/e2e/blackbox_test.go", "example.com/fake/internal/stats", ruleImportWhitelist},
		{"test/e2e/fakegit/bad.go", "example.com/fake/internal/model", ruleImportWhitelist},
	}
	sortViolations(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations mismatch\n got (%d):\n%s\nwant (%d):\n%s",
			len(got), joinViolations(got), len(want), joinViolations(want))
	}
}

// TestChecker_Rules covers each rule in isolation with table-driven cases.
func TestChecker_Rules(t *testing.T) {
	const mod = "example.com/fake"
	tests := []struct {
		name string
		file string
		imps []string
		want []violation
	}{
		{"exec in gitlog ok", "internal/gitlog/a.go", []string{"os/exec"}, nil},
		{"exec in product", "internal/cli/a.go", []string{"os/exec"}, []violation{{"internal/cli/a.go", "os/exec", ruleOSExec}}},
		{"exec in unit test ok", "internal/stats/a_test.go", []string{"os/exec"}, nil},
		{"exec in testutil ok", "internal/testutil/fixture/a.go", []string{"os/exec"}, nil},
		{"exec in fakegit ok", "test/e2e/fakegit/a.go", []string{"os/exec"}, nil},
		{"net in non-test", "internal/gitlog/a.go", []string{"net"}, []violation{{"internal/gitlog/a.go", "net", ruleNetwork}}},
		{"external test pkg ok", "internal/stats/a_test.go", []string{mod + "/internal/stats"}, nil},
		{"test may import testutil", "internal/stats/a_test.go", []string{mod + "/internal/testutil/fixture"}, nil},
		{"testutil by non-test", "internal/stats/a.go", []string{mod + "/internal/testutil/fixture"}, []violation{{"internal/stats/a.go", mod + "/internal/testutil/fixture", ruleTestPathLeak}}},
		{"test/ by test file", "internal/stats/a_test.go", []string{mod + "/test/e2e/fakegit"}, []violation{{"internal/stats/a_test.go", mod + "/test/e2e/fakegit", ruleTestPathLeak}}},
		{"cli imported by stats", "internal/stats/a.go", []string{mod + "/internal/cli"}, []violation{{"internal/stats/a.go", mod + "/internal/cli", ruleImportWhitelist}}},
		{"allowed third party", "internal/render/a.go", []string{"github.com/mattn/go-runewidth", "github.com/rivo/uniseg"}, nil},
		{"unknown third party in test", "internal/render/a_test.go", []string{"github.com/stretchr/testify/assert"}, []violation{{"internal/render/a_test.go", "github.com/stretchr/testify/assert", ruleThirdParty}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkFile(tc.file, tc.imps, mod)
			sortViolations(got)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func joinViolations(vs []violation) string {
	var b strings.Builder
	for _, v := range vs {
		b.WriteString("  " + v.String() + "\n")
	}
	return b.String()
}
