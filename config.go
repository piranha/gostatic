// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"fmt"
	"strings"
	"regexp"
)

type RuleDesc struct {
	Deps []string
	Rules RuleList
}

type SiteConfig struct {
	Templates []string
	Source string
	Output string
	Rules map[string]RuleDesc
}



func NewSiteConfig(source string) *SiteConfig {
	cfg := &SiteConfig{Rules: make(map[string]RuleDesc)}

	indent := ""
	level := 0
	prefix := regexp.MustCompile("^[ \t]*")

	for _, line := range strings.Split(source, "\n") {
		// check indent
		indnew := prefix.FindString(line)
		switch {
		case len(indnew) > len(indent):
			level += 1
		case len(indnew) < len(indent):
			level -= 1
		}
		indent = indnew

		fmt.Printf("%d: %s\n", level, line)
	}

	return cfg
}
