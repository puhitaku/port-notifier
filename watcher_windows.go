//go:build windows

package main

import (
	"errors"
	"fmt"
	"syscall"

	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sys/windows/registry"
)

func listPorts() (mapset.Set[string], error) {
	ports := mapset.NewSet[string]()

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, "HARDWARE\\DEVICEMAP\\SERIALCOMM", registry.READ)
	if err != nil {
		// SERIALCOMM will not exist if no COM port is present on the system
		if errors.Is(err, syscall.ENOENT) {
			return ports, nil
		}
		return nil, fmt.Errorf("failed to open a registry key: %w", err)
	}
	defer key.Close()

	names, _ := key.ReadValueNames(0)
	for _, n := range names {
		if v, _, err := key.GetStringValue(n); err == nil {
			ports.Add(v)
		}
	}
	return ports, nil
}
