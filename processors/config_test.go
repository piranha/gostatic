package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
	"testing"
)

// Tests if CRLF EOL characters in the page content are properly
// processed by ProcessConfig.
func TestProcessConfigCRLF(t *testing.T) {
	page := &gostatic.Page{}

	pageText := "The actual page content.\r\n"
	pageText += "----\r\n"
	pageText += "Some dashes in the page content itself.\r\n"

	content := "title: Test Page\r\n"
	content += "date: 2015-01-01\r\n"
	content += "salve: hello\r\n"
	content += "munde: world\r\n"
	content += "----\r\n"
	content += pageText

	page.SetContent(content)

	ProcessConfig(page, nil)

	checkPage(t, pageText, page)
}

// Tests if LF EOL characters in the page content are properly
// processed by ProcessConfig.
func TestProcessConfigLF(t *testing.T) {
	page := &gostatic.Page{}

	pageText := "The actual page content.\n"
	pageText += "----\n"
	pageText += "Some dashes in the page content itself.\n"

	content := "title: Test Page\n"
	content += "date: 2015-01-01\n"
	content += "salve: hello\n"
	content += "munde: world\n"
	content += "----\n"
	content += pageText

	page.SetContent(content)

	ProcessConfig(page, nil)

	checkPage(t, pageText, page)
}

// Runs some checks against the parsed page.
func checkPage(t *testing.T, pageText string, page *gostatic.Page) {
	if page.Content() != pageText {
		t.Errorf("expected '%s', got '%s'", pageText, page.Content())
	}

	if len(page.PageHeader.Other) != 2 {
		t.Errorf("expected length of 2, got '%d'", len(page.PageHeader.Other))
	}

	// Note: gostatic lib capitalizes the properties for some reason. So it
	// converted 'salve' to 'Salve' etc.
	salve := page.PageHeader.Other["Salve"]
	if salve != "hello" {
		t.Errorf("expected 'hello', got '%s'", salve)
	}

	munde := page.PageHeader.Other["Munde"]
	if munde != "world" {
		t.Errorf("expected 'world', got '%s'", munde)
	}
}
