title = Blog

----

This is a list of posts on my blog:

{{ range .Site.Pages.Children .Url }}
 - [{{ .Title }}]({{ .Url }})
{{ end }}
