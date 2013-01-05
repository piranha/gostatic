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
	Source    string
	Output    string
	Rules     RuleMap
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

func NewSiteConfig(path string) (*SiteConfig, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	basepath, _ := filepath.Split(path)
	cfg := &SiteConfig{Rules: make(RuleMap)}

	indent := 0
	level := 0
	prefix := regexp.MustCompile("^[ \t]*")
	comment := regexp.MustCompile(`[^\\]#`)

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
		cfg.Templates = strings.Split(bits[1], " ")
		for i, template := range cfg.Templates {
			cfg.Templates[i] = filepath.Join(base, template)
		}
	case "SOURCE":
		cfg.Source = filepath.Join(base, bits[1])
	case "OUTPUT":
		cfg.Output = filepath.Join(base, bits[1])
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

func (cmd Command) MatchesAny(prefixes CommandList) bool {
	for _, prefix := range prefixes {
		if cmd.Matches(prefix) {
			return true
		}
	}
	return false
}

func (commands CommandList) MatchedIndex(prefix Command) int {
	for i, cmd := range commands {
		if cmd.Matches(prefix) {
			return i
		}
	}
	return -1
}

func (rule Rule) MatchedCommand(prefix Command) *Command {
	i := rule.Commands.MatchedIndex(prefix)
	if i == -1 {
		return nil
	}

	return &rule.Commands[i]
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
		if !matched {
			matched, err = filepath.Match(pat, name)
		}
		errhandle(err)
		if matched {
			return pat, rule
		}
	}

	return "", nil
}
