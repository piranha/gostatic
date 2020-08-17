package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
)

// DefaultProcessors is variable of processors
var DefaultProcessors = gostatic.ProcessorMap{
	"template":               NewTemplateProcessor(),
	"inner-template":         NewInnerTemplateProcessor(),
	"config":                 NewConfigProcessor(),
	"markdown":               NewMarkdownProcessor(),
	"ext":                    NewExtProcessor(),
	"directorify":            NewDirectorifyProcessor(),
	"tags":                   NewTagsProcessor(),
	"paginate":               NewPaginateProcessor(),
	"paginate-collect-pages": NewPaginateCollectPagesProcessor(),
	"relativize":             NewRelativizeProcessor(),
	"rename":                 NewRenameProcessor(),
	"external":               NewExternalProcessor(),
	"ignore":                 NewIgnoreProcessor(),
	"ignorefuture":           NewIgnoreFutureProcessor(),
	"jekyllify":              NewJekyllifyProcessor(),
	"yaml":                   NewYamlProcessor(),
	"minify":                 NewMinifyProcessor(),
}
