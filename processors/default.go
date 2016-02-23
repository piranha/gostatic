package processors

import (
	gostatic "github.com/piranha/gostatic/lib"
)

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
}
