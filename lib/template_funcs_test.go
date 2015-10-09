package gostatic

import (
	"testing"
)

func TestExcerpt(t *testing.T) {
	inputText := "The quick'ned brown fox, jumps; over the lazy doo-dawg."
	var testTable = []struct {
		maxWords int
		expected string
	}{
		{0, ""},
		{1, "The [...]"},
		{3, "The quick'ned brown [...]"},
		{4, "The quick'ned brown fox, [...]"},
		{8, "The quick'ned brown fox, jumps; over the lazy [...]"},
		{99, "The quick'ned brown fox, jumps; over the lazy doo-dawg."},
	}

	for _, s := range testTable {
		out := Excerpt(inputText, s.maxWords)
		if out != s.expected {
			t.Errorf("Expected \"%s\", got \"%s\"", s.expected, out)
		}
	}
}
