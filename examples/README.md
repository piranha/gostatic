# Gostatic Examples

Those dirs contain solutions to a question "How do I do..." people ask
sometimes. I should've started collecting them earlier, but better later than
never I guess. :)

## `single-tag-page` 

If you want to render a single page containing all tags. Idea is that tags are
links to a regular pages, gostatic core has almost no idea about them except for
`Site.WithTag` method. 

So what we do is `ignore` tag pages, and instead create a template which
iterates over `.Site.Pages.Children "tags/*"` pages and then uses their `.Title`
as a tag id.
