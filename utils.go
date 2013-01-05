// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	bf "github.com/russross/blackfriday"
	"log"
	"os"
	"fmt"
)

func errhandle(err error) {
	if err == nil {
		return
	}
	panic(err)
	log.Fatalf("ERR %s\n", err)
	os.Exit(1)
}

func out(format string, args... interface{}) {
	fmt.Printf(format, args...)
}

func debug(format string, args...  interface{}) {
	if !*verbose {
		return
	}
	fmt.Printf(format, args...)
}

func SliceStringIndexOf(haystack []string, needle string) int {
	for i, elem := range haystack {
		if elem == needle {
			return i
		}
	}
	return -1
}

func Markdown(source string) string {
	// set up the HTML renderer
	flags := 0
	flags |= bf.HTML_USE_SMARTYPANTS
	flags |= bf.HTML_SMARTYPANTS_FRACTIONS
	renderer := bf.HtmlRenderer(flags, "", "")

	// set up the parser
	ext := 0
	ext |= bf.EXTENSION_NO_INTRA_EMPHASIS
	ext |= bf.EXTENSION_TABLES
	ext |= bf.EXTENSION_FENCED_CODE
	ext |= bf.EXTENSION_AUTOLINK
	ext |= bf.EXTENSION_STRIKETHROUGH
	ext |= bf.EXTENSION_LAX_HTML_BLOCKS
	ext |= bf.EXTENSION_SPACE_HEADERS

	return string(bf.Markdown([]byte(source), renderer, ext))
}
