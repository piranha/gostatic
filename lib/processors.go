package gostatic

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// Preprocessor is a processor which will be executed during initialization
	// stage
	Pre = 1 << iota
	Hidden
	Post
)

type Processor interface {
	Process(page *Page, args []string) error
	Description() string
	Mode() int
}

type ProcessorMap map[string]Processor

func (pm ProcessorMap) ProcessorSummary() {
	keys := make([]string, 0, len(pm))
	for k := range pm {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		p := pm[k]
		if p.Mode()&Hidden != 0 {
			continue
		}
		pre := ""
		if p.Mode()&Pre != 0 {
			pre = "(preprocessor)"
		}
		fmt.Printf("- %s %s\n\t%s\n", k, pre, p.Description())
	}
}

func (cmd *Command) Name() string {
	c := string(*cmd)
	if strings.HasPrefix(c, ":") {
		return "external"
	} else {
		return strings.SplitN(c, " ", 2)[0]
	}
}

func (cmd *Command) Args() []string {
	c := string(*cmd)
	if strings.HasPrefix(c, ":") {
		return strings.Split(c[1:], " ")
	} else {
		return strings.Split(c, " ")[1:]
	}
}


func (cmd *Command) Processor(s *Site) (Processor, error) {
	name := cmd.Name()
	processor := s.Processors[name]
	if processor == nil {
		return nil, fmt.Errorf("processor '%s' not found", name)
	}
	return processor, nil
}

func (cmd *Command) IsPre(s *Site) (bool, error) {
	processor, err := cmd.Processor(s)
	if err != nil {
		return false, err
	}
	return processor.Mode()&Pre != 0, nil
}


func (s *Site) ProcessCommand(page *Page, cmd *Command, pre bool) error {
	processor, err := cmd.Processor(s)
	if err != nil {
		return err
	}
	if (processor.Mode()&Pre != 0) != pre {
		return nil
	}
	return processor.Process(page, cmd.Args())
}
