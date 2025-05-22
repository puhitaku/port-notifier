//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/deckarep/golang-set/v2"
)

func listPorts() (mapset.Set[string], error) {
	ports := mapset.NewSet[string]()

	entries, err := os.ReadDir("/dev")
	if err != nil {
		return ports, err
	}

	filter := func(name string) bool {
		if runtime.GOOS == "darwin" {
			return strings.HasPrefix(name, "cu")
		}
		return strings.HasPrefix(name, "tty")
	}

	for _, e := range entries {
		n := e.Name()
		if filter(n) {
			ports.Add(filepath.Join("/dev", n))
		}
	}
	return ports, nil
}
