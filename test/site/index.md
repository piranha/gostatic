title: Alexander Solovyov
stuff: one, two, three

----

## Online presence

 - [blog](blog/)

## Stuff

{{ range split .Other.Stuff "," }}
  - {{ . }}
{{ end }}

{{ exec "echo" "hello" }}

# Hello

* hi
```
# hi
```
* ho
