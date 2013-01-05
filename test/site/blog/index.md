title: Blog

----

This is a list of posts on my blog:

{{ range .Site.Pages.Children .Url }}
{{ if HasChanged "year" .Date.Year }}
 - Year {{ .Date.Year }}
{{ end }}
 - [{{ .Title }}]({{ $.UrlTo . }})
{{ end }}
