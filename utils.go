// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"log"
	"os"
)

func errhandle(err error) {
	if err == nil {
		return
	}
	panic(err)
	log.Fatalf("ERR %s\n", err)
	os.Exit(1)
}

func SliceStringIndexOf(haystack []string, needle string) int {
	for i, elem := range haystack {
		if elem == needle {
			return i
		}
	}
	return -1
}
