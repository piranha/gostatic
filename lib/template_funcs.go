package gostatic

import (
	"fmt"
	"hash/adler32"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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
	if bloc == nil {
		return "", nil
	}
	value = value[bloc[1]:]

	eloc := ere.FindIndex([]byte(value))
	if eloc == nil {
		return "", nil
	}

	return value[:eloc[0]], nil
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

// Matches returns tre if regexp `pattern` matches string `value'.
func Matches(pattern, value string) (bool, error) {
	return regexp.MatchString(pattern, value)
}

func ReFind(pattern, value string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	m := re.FindStringSubmatch(value)

	switch len(m) {
	case 0:
		return "", nil
		// return first submatch if there is any
	case 1:
		return m[0], nil
	default:
		return m[1], nil
	}
}

// Exec runs a `cmd` with all supplied arguments
func Exec(cmd string, arg ...string) (string, error) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		current, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path, err = exec.LookPath(filepath.Join(current, cmd))
		if err != nil {
			return "", fmt.Errorf("command '%s' not found", cmd)
		}
	}

	c := exec.Command(path, arg...)
	out, err := c.CombinedOutput()
	return string(out), err
}

// ExecText runs a `cmd` with all supplied arguments with stdin bound to last argument
func ExecText(cmd string, arg ...string) (string, error) {
	text := arg[len(arg)-1]
	arg = arg[:len(arg)-1]

	path, err := exec.LookPath(cmd)
	if err != nil {
		current, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path, err = exec.LookPath(filepath.Join(current, cmd))
		if err != nil {
			return "", fmt.Errorf("command '%s' not found", cmd)
		}
	}

	c := exec.Command(path, arg...)
	stdin, err := c.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, text)
	}()

	out, err := c.CombinedOutput()
	return string(out), err
}

// Excerpt takes an input string (for example, text from a blog post), and
// truncates it to the amount of words given in maxWords. For instance, given
// the text:
//
//	"The quick brown fox jumps, over the lazy dog."
//
// and the given maxWords of 0, 1, 3, 4, and 6, 999, it will return in order:
//
//	"" // an empty string
//	"The [...]"
//	"The quick brown [...]"
//	"The quick brown fox [...]"
//	"The quick brown fox jumps, over the lazy dog."
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
	return value%2 == 0
}

func Odd(value int) bool {
	return !Even(value) // checking for 1 fails for negative numbers
}

func Count(text string) int {
	return len(strings.Split(text, " "))
}

func ReadingTime(text string) int {
	return (Count(text) + 199) / 200
}

func Some(strs ...interface{}) string {
	for _, x := range strs {
		switch v := x.(type) {
		case nil:
			continue
		case string:
			if v != "" {
				return v
			}
		default:
			s := fmt.Sprintf("%v", v)
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func Dir(value string) string {
	return filepath.Dir(value)
}

func Base(value string) string {
	return filepath.Base(value)
}

func TemplateMarkdown(strs ...string) string {
	return Markdown(strs[len(strs)-1], strs[:len(strs)-1])
}

func Absurl(prefix, path string) (string, error) {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return path, nil
	}
	base, err := url.Parse(prefix)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	res := base.ResolveReference(u)
	return res.String(), nil
}

func AbcSort(pages PageSlice) PageSlice {
	sort.SliceStable(pages, func(i, j int) bool { return pages[i].Name() < pages[j].Name() })
	return pages
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
	"trim":           strings.TrimSpace,
	"replace":        Replace,
	"replacen":       ReplaceN,
	"replacere":      ReplaceRe,
	"split":          Split,
	"contains":       Contains,
	"starts":         Starts,
	"ends":           Ends,
	"matches":        Matches,
	"refind":         ReFind,
	"markdown":       TemplateMarkdown,
	"exec":           Exec,
	"exectext":       ExecText,
	"excerpt":        Excerpt,
	"even":           Even,
	"odd":            Odd,
	"count":          Count,
	"reading_time":   ReadingTime,
	"some":           Some,
	"dir":            Dir,
	"base":           Base,
	"absurl":         Absurl,
	"abcsort":        AbcSort,
}
