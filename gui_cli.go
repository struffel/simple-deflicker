//go:build darwin || cli
// +build darwin cli

package main

import "errors"

func startGUI() error {
	return errors.New("GUI mode is not available in this build. Use -source and -destination to run in CLI mode")
}
