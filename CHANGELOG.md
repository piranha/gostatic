# Gostatic changelog

## 2.33

- New template function `absurl` to join urls sanely

## 2.32

- Template function `markdown` broke backward compatibility, fixed now (always wanted more than one argument)

## 2.31

- Built-in support for [chroma](https://github.com/alecthomas/chroma) highlighter (see README)
- Hot reloading reconnects websocket if it closes, so it persists between runs of gostatic

## 2.30

- Hot reloading will now dispatch JS event `hotreload` on window
- Mac binary is now universal (x64 and arm64 simultaneously)

## 2.29

- added `trim` template function

## 2.28

Hot reloading now uses morphdom and because of that screen doesn't flicker upon change

## 2.27

- added `datefilename` processor

## 2.26

- added ability to parse Github-like frontmatter, i.e. `---\nvar: content\n---`

## 2.25

- all glob matches (like in config) support double stars to descend recursively, like `blog/**/*.md`
- `dir` and `base` functions to manipulate paths in templates
- `refind` function to find strings inside strings

## 2.24

- `cut` now will return empty string if one of regexes did not match
- now invalid processor names from config are reported in a better way - their name is included in error message

## 2.23

Enabled unsafe HTML for goldmark - it won't omit raw HTML when rendering

## 2.22

Switch markdown library from [blackfriday](github.com/russross/blackfriday) to [goldmark](https://github.com/yuin/goldmark/), adding support of Commonmark, and, notably, smartypants-like features.

## 2.21

New template function - `some`: returns first non-nil value, intended to use instead of lengthy ifs.

## 2.20

Increase number in `version.go`, because it still was 2.17. :)

## 2.19

- new template functions: `count` and `reading_time`
- hot reloading has an exponential timeout up to a second to reduce flickering

## 2.18

Hot HTML code reload when in dev mode (using `gostatic -w/--watch`).

## 2.17

`.Has` for pages, `.Where` and `.WhereNot` for page lists.

## 2.16

`exectext` function.

## 2.15

`.Reverse` is now available as a method on page lists.

## 2.14

- new template function: `matches`, checks for regexp in a string.
- fixed parsing tags in YAML header
- inner templates report their errors better now
- support for BOM (easier to use with files created on Windows)
- CRLF support
- also, gomod - we have pinned versions of dependencies

## 2.13

New processors for people switching from Jekyll: `jekyllify` to convert posts to
a familiar path, and `yaml` to process headers as YAML (rather than whatever
custom stuff gostatic uses by default).

## 2.12

Now `cut` searches for the `end` *after* end of `begin` match.

## 2.11

New template function: `replacere`.

## 2.10

Two new template functions: `even` and `odd`.

## 2.9

`gostatic -w` now waits 10 ms before doing anything to prevent problems with
emacs-style file changes, when it first creates empty file in place of an old
one and then moves changes over to it.

## 2.8

Two new template functions: `starts` and `ends`.

## 2.7

Ability to have multiple configurations for a single path (so you can have
multiple outputs from one file).

## 2.6

Sort pages with same date alphabetically.

## 2.5

Get `exec` template function back.

## 2.4

- Fixed handling \r\n in `config` processor
- Now errors of `external` processor are propagated and you'll see them

## 2.3

gostatic is now a [library](https://github.com/piranha/gostatic#extensibility) (thanks @zhuharev)! Plus:

- `exec` function in templates
- `exceprt` function in templates (thanks @krpors)
- gostatic no longer fails on vim's temp files (thanks @krpors)

## 2.2

Make example site (gostatic -i) work with current gostatic.

## 2.1

Fix `rename` processor for Windows.

## 2.0

Major version - **breaking changes**.

- **Backward incompatible** - [template functions](https://github.com/piranha/gostatic#global-functions) `cut` and `split` now have different order of arguments to better support [template pipelining](https://golang.org/pkg/text/template/#hdr-Pipelines).
- Pagination is now supported, see `paginate` [processor](https://github.com/piranha/gostatic#processors) and `paginator` [template function](https://github.com/piranha/gostatic#global-functions).
- Template and config changes are now tracked and will result in full re-render.
- [Page](https://github.com/piranha/gostatic#page-interface) now has `.Raw` property, containing unprocessed data (but after `config` being consumed).
- `strip_newlines`, `replace`, `replacen`, `contains`, `markdown` [template functions](https://github.com/piranha/gostatic#global-functions).
- [Page list](https://github.com/piranha/gostatic#page-list-interface) new methods: `.Slice` and `.GlobSource`.

## 1.17

More fsnotify stuff.

## 1.16

Updated fsnotify; potentially better watcher behavior.

## 1.15

Ability to specify folders with templates.

## 1.14

"split" function in templates to generate array from string.

## 1.13

Ability to hide pages with `hide: true` in page header.

## 1.12

More functions for templates: `truncate` and `split_html`.

## 1.11

Make errors when processing template at least a bit better.

## 1.10

More strict split in ProcessConfig.

## 1.9

Enable header ids in markdown processing.

## 1.8

Somewhat simplified watch code and it started working.

## 1.7

- Enable footnotes
- Smaller binaries (by skipping debug info)
- Fix directory walking error handling
- Watch source directory instead of destination
- Watch templates for changes

## 1.6

Fixed crash on `PageSlice.Prev` when no previous pages exist.

## 1.5

Add `PageSlice.Next` and `PageSlice.Prev`

## 1.4

Ability to print page metadata as json (`gostatic --dump src/path/to/url config`).

## 1.3

- Fixed bug with empty bodies
- Ability to have comments in page header (with `#`)

## 1.2

- Fix `PageSlice.Slice` crashes
- Fix `cut` to not fail when search returns no results
- `Page.UrlMatches`
- Compare tag pages by path (not by title)

## 1.1

- Fix example site to escape entities
- Fix symlink handling

## 1.0

First tagged release, lots of good stuff. :)
