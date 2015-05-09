// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type Command string

type CommandList []Command

type Rule struct {
	Deps     []string
	Commands CommandList
}

type RuleMap map[string]*Rule

type SiteConfig struct {
	Templates []string
	Base      string
	Source    string
	Output    string
	Rules     RuleMap
	Other     map[string]string
}

func NewSiteConfig(path string) (*SiteConfig, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	basepath, _ := filepath.Split(path)
	cfg := &SiteConfig{
		Rules: make(RuleMap),
		Other: make(map[string]string),
		Base:  basepath,
	}

	indent := 0
	level := 0
	prefix := regexp.MustCompile("^[ \t]*")
	comment := regexp.MustCompile(`(^|[^\\])#`)

	var current *Rule

	for i, line := range strings.Split(string(source), "\n") {
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
		commentloc := comment.FindIndex([]byte(line))
		if commentloc != nil {
			line = line[:commentloc[0]]
		}
		line = strings.Replace(line, "\\#", "#", -1)
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		// is this a constant declaration?
		if level == 0 && strings.Index(line, "=") != -1 {
			cfg.ParseVariable(basepath, line)
			continue
		}

		// not a constant, then a Rule start?
		if level == 0 {
			current = cfg.ParseRule(line)
			continue
		}

		if level == 1 {
			if current == nil {
				return nil, fmt.Errorf("Indent without rules, line %d", i+1)
			}
			current.ParseCommand(line)
			continue
		}

		return nil, fmt.Errorf("Unhandled situation on line %d: %s",
			i+1, line)
	}

	return cfg, nil
}

// *** Parsing methods

func (cfg *SiteConfig) ParseVariable(base string, line string) {
	bits := TrimSplitN(line, "=", 2)
	switch bits[0] {
	case "TEMPLATES":
		templates := strings.Split(bits[1], " ")
		for _, template := range templates {
			path := filepath.Join(base, template)
			isDir, err := IsDir(path)

			if err != nil {
				errexit(fmt.Errorf("Template does not exist: %s", err))
			}

			if isDir {
				files, _ := filepath.Glob(filepath.Join(path, "*.tmpl"))
				for _, fn := range files {
					cfg.Templates = append(cfg.Templates, fn)
				}
			} else {
				cfg.Templates = append(cfg.Templates, path)
			}
		}
	case "SOURCE":
		cfg.Source = filepath.Join(base, bits[1])
	case "OUTPUT":
		cfg.Output = filepath.Join(base, bits[1])
	default:
		cfg.Other[Capitalize(bits[0])] = bits[1]
	}
}

func (cfg *SiteConfig) ParseRule(line string) *Rule {
	bits := TrimSplitN(line, ":", 2)
	deps := NonEmptySplit(bits[1], " ")
	rd := &Rule{
		Deps:     deps,
		Commands: make(CommandList, 0),
	}

	cfg.Rules[bits[0]] = rd

	return rd
}

func (rule *Rule) ParseCommand(line string) {
	rule.Commands = append(rule.Commands, Command(line))
}

// *** Traversing methods

func (cmd Command) Matches(prefix Command) bool {
	return cmd == prefix || strings.HasPrefix(string(cmd), string(prefix)+" ")
}

func (rule *Rule) IsDep(page *Page) bool {
	for _, dep := range rule.Deps {
		matches, err := filepath.Match(dep, page.Source)
		if err == nil && matches {
			return true
		}
	}
	return false
}

func (rules RuleMap) MatchedRule(path string) (string, *Rule) {
	if rules[path] != nil {
		return path, rules[path]
	}

	_, name := filepath.Split(path)
	if rules[name] != nil {
		return name, rules[name]
	}

	for pat, rule := range rules {
		matched, err := filepath.Match(pat, path)
		errhandle(err)
		if matched {
			return pat, rule
		}
	}

	for pat, rule := range rules {
		matched, err := filepath.Match(pat, name)
		errhandle(err)
		if matched {
			return pat, rule
		}
	}

	return "", nil
}
