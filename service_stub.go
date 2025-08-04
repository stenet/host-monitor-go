//go:build !windows

package main

import (
	"fmt"
	"os"
	"time"
)

func isWindowsService() bool {
	return false
}

func runAsWindowsService(debug bool, seqURL string, interval time.Duration) {
	fmt.Fprintf(os.Stderr, "Windows Service Funktionalität ist nur unter Windows verfügbar\n")
	os.Exit(1)
}

func installWindowsService(serviceName, seqURL string, interval time.Duration, debug bool) {
	fmt.Fprintf(os.Stderr, "Windows Service Installation ist nur unter Windows verfügbar\n")
	os.Exit(1)
}

func uninstallWindowsService(serviceName string) {
	fmt.Fprintf(os.Stderr, "Windows Service Deinstallation ist nur unter Windows verfügbar\n")
	os.Exit(1)
}