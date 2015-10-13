// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Command is a command belonging to a Rule. For example, `markdown', `directorify'.
type Command string

// CommandList is a slice of Commands.
type CommandList []Command

// Rule is a collection of a slice of dependencies, with a slice of commands in the
// form of a CommandList.
type Rule struct {
	Deps     []string
	Commands CommandList
}

type RuleMap map[string]*Rule

// SiteConfig contains the data for a complete parsed site configuration file.
type SiteConfig struct {
	Templates []string
	Base      string
	Source    string
	Output    string
	Rules     RuleMap
	Other     map[string]string
	changedAt time.Time
}

// NewSiteConfig parses the given `path' file to a *SiteConfig. Will return a nil
// pointer plus the non-nil error if the parsing has failed.
func NewSiteConfig(path string) (*SiteConfig, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	basepath, _ := filepath.Split(path)
	cfg := &SiteConfig{
		Rules:     make(RuleMap),
		Other:     make(map[string]string),
		Base:      basepath,
		changedAt: stat.ModTime(),
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
			level++
		case indnew < indent:
			level--
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
				return nil, fmt.Errorf("indent without rules, line %d", i+1)
			}
			current.ParseCommand(cfg, line)
			continue
		}

		return nil, fmt.Errorf("unhandled situation on line %d: %s",
			i+1, line)
	}

	return cfg, nil
}

// *** Parsing methods

var VarRe = regexp.MustCompile(`\$\(([^\)]+)\)`)

func (cfg *SiteConfig) SubVars(s string) string {
	return VarRe.ReplaceAllStringFunc(s, func(m string) string {
		name := VarRe.FindStringSubmatch(m)[1]
		switch name {
		case "TEMPLATES":
			return strings.Join(cfg.Templates, ", ")
		case "SOURCE":
			return cfg.Source
		case "OUTPUT":
			return cfg.Output
		default:
			return cfg.Other[Capitalize(name)]
		}
	})
}

func (cfg *SiteConfig) ParseVariable(base string, line string) {
	bits := TrimSplitN(line, "=", 2)
	name := bits[0]
	value := cfg.SubVars(bits[1])

	switch name {
	case "TEMPLATES":
		templates := strings.Split(value, " ")
		for _, template := range templates {
			path := filepath.Join(base, template)
			isDir, err := IsDir(path)

			if err != nil {
				errexit(fmt.Errorf("template does not exist: %s", err))
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
		cfg.Source = filepath.Join(base, value)
	case "OUTPUT":
		cfg.Output = filepath.Join(base, value)
	default:
		cfg.Other[Capitalize(name)] = value
	}
}

func (cfg *SiteConfig) ParseRule(line string) *Rule {
	bits := TrimSplitN(line, ":", 2)
	deps := NonEmptySplit(cfg.SubVars(bits[1]), " ")
	rd := &Rule{
		Deps:     deps,
		Commands: make(CommandList, 0),
	}

	cfg.Rules[bits[0]] = rd

	return rd
}

func (rule *Rule) ParseCommand(cfg *SiteConfig, line string) {
	line = cfg.SubVars(line)
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
