// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"os"
	"log"
)

func errhandle(err error) {
	if err == nil {
		return
	}
	log.Fatalf("ERR %s\n", err)
	os.Exit(1)
}
