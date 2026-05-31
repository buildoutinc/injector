{{ range .Versions }}
{{ range .CommitGroups -}}
## {{ .Title }}

{{ range .Commits -}}
- {{ .Subject }} ({{ .Hash.Short }})
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
## {{ .Title }}

{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}

{{- $unmatched := list -}}
{{- range .Commits -}}
{{- if not .Type -}}
{{- $unmatched = append $unmatched . -}}
{{- end -}}
{{- end -}}
{{- if $unmatched }}
## Other Changes

{{ range $unmatched -}}
- {{ .Subject }} ({{ .Hash.Short }})
{{ end }}
{{- end -}}
{{ end -}}
