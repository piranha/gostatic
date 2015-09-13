package processors

import (
	"bytes"
	"errors"
	"fmt"
	gostatic "github.com/piranha/gostatic/lib"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExternalProcessor struct {
}

func NewExternalProcessor() *ExternalProcessor {
	return &ExternalProcessor{}
}

func (p *ExternalProcessor) Process(page *gostatic.Page, args []string) error {
	return ProcessExternal(page, args)
}

func (p *ExternalProcessor) Description() string {
	return "run external command to process content (shortcut ':')"
}

func (p *ExternalProcessor) Mode() int {
	return 0
}

func ProcessExternal(page *gostatic.Page, args []string) error {
	if len(args) < 1 {
		return errors.New("'external' rule needs a command name")
	}
	cmdName := args[0]
	cmdArgs := args[1:]

	path, err := exec.LookPath(cmdName)
	if err != nil {
		path, err = exec.LookPath(filepath.Join(page.Site.Base, cmdName))
		if err != nil {
			return fmt.Errorf("command '%s' not found", cmdName)
		}
	}

	cmd := exec.Command(path, cmdArgs...)
	cmd.Stdin = strings.NewReader(page.Content())
	cmd.Dir = page.Site.Base
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("'%s' failed: %s\n%s",
			strings.Join(args, " "), err, stderr.String())
	}

	page.SetContent(string(data))
	return nil
}
