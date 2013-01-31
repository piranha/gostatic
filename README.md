# gostatic

Gostatic is a static site generator. What differs it from most of other tools is
that it's written in Go and tracks changes, which means it should work
reasonably [fast](#speed).

Features include:

 - Dependency tracking and re-rendering only changed pages
 - Markdown support
 - Flexible [filter system](#processors)
 - Simple [config syntax](#configuration)
 - HTTP server and watcher (instant rendering on changes)

Binary builds:

 - [Linux](http://solovyov.net/files/gostatic-linux)
 - [OS X](http://solovyov.net/files/gostatic-osx)
 - [Windows](http://solovyov.net/files/gostatic-win.exe)

Examples of use: `test` directory in this repository and
[my site](https://github.com/piranha/solovyov.net).

## Approach

Each given file is processed through a pipeline of filters, which modify the
state and then rendered on disk. Single input file corresponds to a single
output file, but filters can generate virtual input files.

Each file can have dependencies, and will be rendered in case it does not exist,
or its source is newer than output, or one of this is the case for one of its
dependencies.

All read pages are sorted by date, found in their config (explained later) or,
in case of their equality (which also happens when they do not have config), by
modification time.

## Speed

On late 2008 MacBook (2.4 GHz, 8 GB RAM, 5400 rpm HDD) it takes `0.3s` to
generate a site of 250 pages. It costs `0.05s` to check there are no
modifications and `0.1s` to re-render a single changed page (along with index
and tag pages, coming to 77 pages in total).

## External resources

 - Jack Pearkes wrote [Heroku buildpack][] for gostatic and an
   [article about it][].

[Heroku buildpack]: https://github.com/pearkes/heroku-buildpack-gostatic
[article about it]: http://pretengineer.com/post/gostatic-buildpack-for-heroku/

## Configuration

Config syntax is Makefile-inspired with some simplifications, look at the
example:

```Makefile
TEMPLATES = site.tmpl
SOURCE = src
OUTPUT = site

*.md:
    config
    ext .html
    directorify
    tags tags/*.tag
    markdown
    template page

index.md: blog/*.md
    config
    ext .html
    inner-template
    markdown
    template

*.tag: blog/*.md
    ext .html
    directorify
    template tag
    markdown
    template page
```

Here we have constants declaration (first three lines) and then three rules. One
for any markdown file, one specifically for index.md and one for generated tags.

Note: Specific rules override matching rules, but there is no very smart logic
in place and matches comparisons are not strictly defined, so if you have
several matches you could end up with any of them. Though there is order: exact
path match, exact name match, glob path match, glob name match. NOTE: this may
change in future.

Rules consist of path/match, list of dependencies (also paths and matches, the
ones listed after colon) and commands.

Each command consists of a name of processor and (possibly) some
arguments. Arguments are separated by spaces.

Note: if a file has no rules whatsoever, it will be copied to exactly same
location at destination as it was in source without being read into memory. So
heavy images etc shouldn't be a problem.

### Constants

There are three configuration constants. `SOURCE` and `OUTPUT` speak for
themselves, and `TEMPLATES` is a list of files which will be parsed as Go
templates. Each file can contain few templates.

You can also use arbitrary names for constants to
[access later](#site-interface) from templates.

## Page header

Page header is in format `name: value`, for example:

```
title: This is a page
tags: test
date: 2013-01-05
```

Available properties:

- `title` - page title.
- `tags` - list of tags, separated by `,`.
- `date` - page date, could be used for blog. Accepts formats from bigger to
  smaller (from `"2006-01-02 15:04:05 -07"` to `"2006-01-02"`)

You can also define arbitrary properties to access later from template, they
will be treated as a string.

## Processors

You can always check list of available processors with `gostatic --processors`.

- `config` - reads config from content. Config should be in format "name: value"
  and separated by four dashes on empty line (`----`) from content.

- `ignore` - ignore file.

- `rename <new-name>` - rename a file to `new-name`. New name can contain `*`,
  then it will be replaced with whatever `*` captured in path match. Right now
  rename touches **whole** path, so be careful (you may need to include whole
  path in rename pattern) - *this may change in future*.

- `ext <.ext>` - change file extension to a given one (which should be prefixed
  with a dot).

- `directorify` - rename a file from `whatever/name.html` to
  `whatever/name/index.html`.

- `markdown` - process content as Markdown.

- `inner-template` - process content as Go template.

- `template <name>` - pass page to a template named `<name>`.

- `tags <name-pattern>` - generate (if not yet) virtual page for all tags of
  current page. This tag page has path formed by replacing `*` in
  `<name-pattern>` with a tag name.

- `relativize` - change all urls archored at `/` to be relative (i.e. add
  appropriate amount of `../`) so that generated content can be deployed in a
  subfolder of a site.

- `external <command> <args...>` - call external command with content of a page
  as stdin and using stdout as a new content of a page.

## Templating

Templating is provided using
[Go templates](http://golang.org/pkg/text/template/). See link for documentation
on syntax.

Each template is executed in context of a page. This means it has certain
properties and methods it can output or call to generate content, i.e. `{{
.Content }}` will output page content in place.

### Global functions

Go template system provides some convenient
[functions](http://golang.org/pkg/text/template/#hdr-Functions), and gostatic
expands on that a bit:

 - `changed <name> <value>` - checks if value has changed since previous call
   with the same name. Storage, used for checking, is global over whole run of
   gostatic, so choose unique names.

 - `cut <value> <begin> <end>` - cut partial content from `<value>`, delimited
   by regular expressions `<begin>` and `<end>`.

 - `hash <value>` - return 32-bit hash of a given value.

 - `version <path>` - return relative url to a page with resulting path `<path>`
   with `?v=<32-bit hash>` appended (used to override cache settings on static
   files).

### Page interface

- `.Site` - global [site object](#site-interface).
- `.Rule` - rule object, matched by page.
- `.Pattern` - pattern, which matched this page.
- `.Deps` - list of pages, which are dependencies for this page.

----

- `.Source` - relative path to page source.
- `.FullSource` - full path to page source.
- `.Path` - relative path to page destination.
- `.FullPath` - full path to page destination.
- `.ModTime` - page last modification time.

----

- `.Title` - page title.
- `.Tags` - list of page tags.
- `.Date` - page date, as defined in [page header](#page-header).
- `.Other` - map of all other properties from [page header](#page-header).

----

- `.Content` - page content.
- `.Url` - page url (i.e. `.Path`, but with `index.html` stripped from the end).
- `.UrlTo <other-page>` - relative url from current to some other page.
- `.Rel <url>` - relative url to given absolute (anchored at `/`) url.

### Page list interface

- `.Get <n>` - [page](#page-interface) number `<n>`.
- `.First` - first page.
- `.Last` - last page.
- `.Len` - length of page list.

----

- `.Children <prefix>` - list of pages, nested under `<prefix>`.
- `.WithTag <tag-name>` - list of pages, tagged with `<tag-name>`.
- `.HasPage <page>` - checks if page list contains a `<page>`.

----

- `.BySource <path>` - finds a page with source path `<path>`.
- `.ByPath <path>` - finds a page with resulting path `<path>`.

### Site interface

- `.Pages` - [list of all pages](#page-list-interface).
- `.Source` - path to site source.
- `.Output` - path to site destination.
- `.Templates` - list of template files used for the site.
- `.Other` - any other properties defined in site config.
