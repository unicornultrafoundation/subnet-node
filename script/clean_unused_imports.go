package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// regex to match import aliases
var importAlias = regexp.MustCompile(`(?m)^\s*([a-zA-Z0-9_]+)\s+\"([^"]+)\"`)

func main() {
	root := "../proto/subnet/k8s"
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".pb.go") {
			return nil
		}
		cleanFile(path)
		return nil
	})
}

func cleanFile(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
		return
	}
	lines := strings.Split(string(data), "\n")
	imports := map[string]int{} // alias -> line number
	for i, line := range lines {
		if m := importAlias.FindStringSubmatch(line); m != nil {
			alias := m[1]
			if alias != "_" && alias != "." && alias != "math" && alias != "fmt" && alias != "proto" {
				imports[alias] = i
			}
		}
	}
	// Scan for usage
	used := map[string]bool{}
	for i, line := range lines {
		if i >= len(lines) || strings.HasPrefix(strings.TrimSpace(line), "import (") {
			continue
		}
		for alias := range imports {
			if strings.Contains(line, alias+".") {
				used[alias] = true
			}
		}
	}
	// Remove unused
	changed := false
	for alias, idx := range imports {
		if !used[alias] {
			lines[idx] = "" // remove line
			changed = true
		}
	}
	if changed {
		out := strings.Join(lines, "\n")
		if err := ioutil.WriteFile(path, []byte(out), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", path, err)
		} else {
			fmt.Printf("Cleaned unused imports in %s\n", path)
		}
	}
}
