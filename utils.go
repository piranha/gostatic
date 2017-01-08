// (c) 2012 Alexander Solovyov
// under terms of ISC license

// Error handling utilities are almost duplicated in lib/utils.go so that they
// don't need to be exported

package main

import (
	"fmt"
	"os"
)

func errhandle(err error) {
	if err == nil {
		return
	}
	fmt.Printf("Error: %s\n", err)
}

func errexit(err error) {
	if err == nil {
		return
	}
	fmt.Printf("Fatal error: %s\n", err)
	os.Exit(1)
}

func out(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func debug(format string, args ...interface{}) {
	if !opts.Verbose {
		return
	}
	fmt.Printf(format, args...)
	os.Stdout.Sync()
}
