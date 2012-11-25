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
	Rules map[string]*RuleDesc
}


func TrimSplitN(s string, sep string, n int) []string {
	bits := strings.SplitN(s, sep, n)
	for i, bit := range bits {
		bits[i] = strings.TrimSpace(bit)
	}
	return bits
}

func NonEmptySplit(s string, sep string) []string {
	bits := strings.Split(s, sep)
	out := make([]string, 0)
	for _, x := range bits {
		if len(x) != 0 {
			out = append(out, x)
		}
	}
	return out
}


func NewSiteConfig(source string) (*SiteConfig, error) {
	cfg := &SiteConfig{Rules: make(map[string]*RuleDesc)}

	indent := 0
	level := 0
	prefix := regexp.MustCompile("^[ \t]*")

	var current *RuleDesc

	for i, line := range strings.Split(source, "\n") {
		// check indent
		indnew := len(prefix.FindString(line))
		switch {
		case indnew > indent:
			level += 1
		case indnew < indent:
			level -= 1
		}
		indent = indnew

		// remove useless stuff from line
		line = line[indent:]
		comment := strings.Index(line, "#")
		if comment != -1 {
			line = line[:comment]
		}
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		// is this a constant declaration?
		if level == 0 && strings.Index(line, "=") != -1 {
			cfg.ParseVariable(line)
			continue
		}

		// not a constant, then a RuleDesc start?
		if level == 0 {
			current = cfg.ParseRuleDesc(line)
			continue
		}

		if level == 1 {
			if current == nil {
				return nil, fmt.Errorf("Indent without rules, line %d", i + 1)
			}
			current.ParseRule(line)
			continue
		}

		return nil, fmt.Errorf("Unhandled situation on line %d: %s",
			i + 1, line)
	}

	return cfg, nil
}


func (cfg *SiteConfig) ParseVariable(line string) {
	bits := TrimSplitN(line, "=", 2)
	switch bits[0] {
	case "TEMPLATES":
		cfg.Templates = strings.Split(bits[1], " ")
	case "SOURCE":
		cfg.Source = bits[1]
	case "OUTPUT":
		cfg.Output = bits[1]
	}
}


func (cfg *SiteConfig) ParseRuleDesc(line string) *RuleDesc {
	bits := TrimSplitN(line, ":", 2)
	deps := NonEmptySplit(bits[1], " ")
	rd := &RuleDesc{
		Deps: deps,
		Rules: make(RuleList, 0),
	}

	cfg.Rules[bits[0]] = rd

	return rd
}


func (rd *RuleDesc) ParseRule(line string) {
	rd.Rules = append(rd.Rules, line)
}
