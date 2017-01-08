// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"fmt"
	bf "github.com/russross/blackfriday"
	"io"
	"os"
	"strings"
	"unicode"
)

var (
	DEBUG bool = false
)

func errhandle(err error) {
	if err == nil {
		return
	}
	fmt.Printf("Error: %s\n", err)
}

func errexit(err error) {
	if err == nil {
		return
	}
	fmt.Printf("Fatal error: %s\n", err)
	os.Exit(1)
}

func out(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func debug(format string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(format, args...)
		os.Stdout.Sync()
	}
}

func drainchannel(out chan string) {
	for {
		select {
		case <-out:
		default:
			return
		}
	}
}

func IsDir(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}

	stat, err := file.Stat()
	if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}

func SliceStringIndexOf(haystack []string, needle string) int {
	for i, elem := range haystack {
		if elem == needle {
			return i
		}
	}
	return -1
}

func Markdown(source string) string {
	// set up the HTML renderer
	flags := 0
	flags |= bf.HTML_USE_SMARTYPANTS
	flags |= bf.HTML_SMARTYPANTS_FRACTIONS
	renderer := bf.HtmlRenderer(flags, "", "")

	// set up the parser
	ext := 0
	ext |= bf.EXTENSION_NO_INTRA_EMPHASIS
	ext |= bf.EXTENSION_TABLES
	ext |= bf.EXTENSION_FENCED_CODE
	ext |= bf.EXTENSION_AUTOLINK
	ext |= bf.EXTENSION_STRIKETHROUGH
	ext |= bf.EXTENSION_SPACE_HEADERS
	ext |= bf.EXTENSION_FOOTNOTES
	ext |= bf.EXTENSION_HEADER_IDS

	return string(bf.Markdown([]byte(source), renderer, ext))
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

func Capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + strings.Map(unicode.ToLower, s[1:])
}

func CopyFile(srcPath, dstPath string) (n int64, err error) {
	fstat, err := os.Lstat(srcPath)

	if err != nil {
		return 0, err
	}

	if fstat.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(srcPath)
		if err != nil {
			return 0, err
		}

		err = os.Symlink(target, dstPath)
		if err != nil {
			return 0, err
		}

		return 1, nil
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	n, err = io.Copy(dst, src)
	return n, err
}
