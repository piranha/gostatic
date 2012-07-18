// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"strings"
	"path/filepath"
)

type RuleList []string
type RuleMap map[string]RuleList

// Return index of a rule which starts with given prefix
func (rules RuleList) MatchedIndex(prefix string) int {
	for i, rule := range rules {
		if rule == prefix || strings.HasPrefix(rule, prefix + " ") {
			return i
		}
	}
	return -1
}

func (rules RuleMap) MatchedRules(path string) (string, []string) {
	if rules[path] != nil {
		return path, rules[path]
	}

	_, name := filepath.Split(path)
	if rules[name] != nil {
		return name, rules[name]
	}

	for pat, rules := range rules {
		matched, err := filepath.Match(pat, name)
		errhandle(err)
		if matched {
			return pat, rules
		}
	}

	return "", nil
}
