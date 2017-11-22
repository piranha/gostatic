package gostatic

import (
	"fmt"
	"hash/adler32"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

var inventory = map[string]interface{}{}

func HasChanged(name string, value interface{}) bool {
	changed := true

	if inventory[name] == value {
		changed = false
	} else {
		inventory[name] = value
	}

	return changed
}

func Cut(begin, end, value string) (string, error) {
	bre, err := regexp.Compile(begin)
	if err != nil {
		return "", err
	}
	ere, err := regexp.Compile(end)
	if err != nil {
		return "", err
	}

	bloc := bre.FindIndex([]byte(value))
	eloc := ere.FindIndex([]byte(value))

	if bloc == nil {
		bloc = []int{0, 0}
	}
	if eloc == nil {
		eloc = []int{len(value)}
	}

	return value[bloc[1]:eloc[0]], nil
}

func Hash(value string) string {
	h := adler32.New()
	io.WriteString(h, value)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Versionize(current *Page, value string) (string, error) {
	page := current.Site.Pages.ByPath(value)
	if page == nil {
		errhandle(fmt.Errorf(
			"trying to versionize page which does not exist: %s, current: %s",
			value, current.Path))
	}
	_, err := page.Process()
	if err != nil {
		return "", err
	}
	c := page.Content()
	h := Hash(c)
	return current.UrlTo(page) + "?v=" + h, nil
}

// Truncate truncates the value string to maximum of the given length, and returns it.
func Truncate(length int, value string) string {
	if length > len(value) {
		length = len(value)
	}
	return value[0:length]
}

// StripHTML removes HTML tags from the value string and returns it.
func StripHTML(value string) string {
	return regexp.MustCompile("<[^>]+>").ReplaceAllString(value, "")
}

// StripNewlines removes all \r and \n characters from the value string,
// and returns it as such.
func StripNewlines(value string) string {
	return regexp.MustCompile("[\r\n]").ReplaceAllString(value, "")
}

// Replace replaces `old' with `new' in the given value string and returns it.
// There is no limit on the amount of replacements.
func Replace(old, new, value string) string {
	return strings.Replace(value, old, new, -1)
}

// ReplaceN replaces the `old' string with the `new' string in the given value,
// n times. If n < 0, there is no limit on the number of replacements.
func ReplaceN(old, new string, n int, value string) string {
	return strings.Replace(value, old, new, n)
}

func ReplaceRe(pattern, repl, value string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(value, repl), nil
}

// Split splits the value using the separator sep, and returns it as a
// string slice.
func Split(sep, value string) []string {
	return strings.Split(value, sep)
}

// Contains returns true if `needle' is contained within `value'.
func Contains(needle, value string) bool {
	return strings.Contains(value, needle)
}

// Starts returns true if `value' starts with `needle'.
func Starts(needle, value string) bool {
	return strings.HasPrefix(value, needle)
}

// Ends returns true if `value' ends with `needle'.
func Ends(needle, value string) bool {
	return strings.HasSuffix(value, needle)
}

// Exec runs a `cmd` with all supplied arguments
func Exec(cmd string, arg ...string) (string, error) {
	c := exec.Command(cmd, arg...)
	out, err := c.CombinedOutput()
	return string(out), err
}

// Excerpt takes an input string (for example, text from a blog post), and
// truncates it to the amount of words given in maxWords. For instance, given
// the text:
//
// 	"The quick brown fox jumps, over the lazy dog."
//
// and the given maxWords of 0, 1, 3, 4, and 6, 999, it will return in order:
//
// 	"" // an empty string
// 	"The [...]"
// 	"The quick brown [...]"
// 	"The quick brown fox [...]"
// 	"The quick brown fox jumps, over the lazy dog."
func Excerpt(text string, maxWords int) string {
	// Unsure who would want this, but still, don't trust them users ;)
	if maxWords <= 0 {
		return ""
	}

	splitup := strings.Split(text, " ")
	if maxWords >= len(splitup) {
		return text
	}
	return strings.Join(splitup[0:maxWords], " ") + " [...]"
}

func Even(value int) bool {
	return value % 2 == 0
}

func Odd(value int) bool {
	return !Even(value) // checking for 1 fails for negative numbers
}

// TemplateFuncMap contains the mapping of function names and their corresponding
// Go functions, to be used within templates.
var TemplateFuncMap = template.FuncMap{
	"changed":        HasChanged,
	"cut":            Cut,
	"hash":           Hash,
	"version":        Versionize,
	"truncate":       Truncate,
	"strip_html":     StripHTML,
	"strip_newlines": StripNewlines,
	"replace":        Replace,
	"replacen":       ReplaceN,
	"replacere":      ReplaceRe,
	"split":          Split,
	"contains":       Contains,
	"starts":         Starts,
	"ends":           Ends,
	"markdown":       Markdown,
	"exec":           Exec,
	"excerpt":        Excerpt,
	"even":           Even,
	"odd":            Odd,
}
