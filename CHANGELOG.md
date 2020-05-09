# Gostatic changelog

## 2.20

Increase number in `version.go`, because it still was 2.17 :)

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
