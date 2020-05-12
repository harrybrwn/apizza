package out

import (
	"io"
	"text/template"

	"github.com/harrybrwn/apizza/pkg/errs"
)

func tmpl(w io.Writer, tmplt string, a interface{}) (err error) {
	t := template.New("apizza")
	t, err = t.Parse(tmplt)
	return errs.Pair(err, t.Execute(w, a))
}

var defaultOrderTmpl = `{{ .OrderName }}
{{- $keycol := .KeyColor -}}
{{- $endcol := .EndColor }}
  {{.KeyColor}}products{{.EndColor}}:{{ range .Products }}
    {{$keycol}}name{{$endcol}}: {{.Name}}
      {{$keycol}}code{{$endcol}}:     {{.Code}}
      {{$keycol}}options{{$endcol}}:{{ range $k, $v := .ReadableOptions }}
         {{$keycol}}{{$k}}{{$endcol}}: {{$v}}{{else}}None{{end}}
      {{$keycol}}quantity{{$endcol}}: {{.Qty}}{{end}}
  {{.KeyColor}}storeID{{.EndColor}}: {{.StoreID}}
  {{.KeyColor}}method{{.EndColor}}:  {{.ServiceMethod}}
  {{.KeyColor}}address{{.EndColor}}: {{.Addr -}}
{{ if .Price }}
  {{.KeyColor}}price{{.EndColor}}:   ${{ .Price -}}
{{else}}{{end}}
`

var cartOrderTmpl = `  {{ .OrderName }} - {{ range .Products }} {{.Code}}, {{end}}
`

var menuCategoryTmpl = ``

var variantTmpl = `{{ .Name }} {{ .Code }}
	price:   {{ .Price }}
	product: {{ .ProductCode }}
`

var shorthandVariantTmpl = `{{ .Name }} {{ .Code }}`

var itemTmpl = `{{.ItemName}}
  Code: {{.ItemCode}}
{{ if .Category }}  Category: {{.Category}}{{else}}{{end}}
`

var productTmpl = `    Description: {{.Description}}
  Variants: {{.Variants}}
  Available sides: {{ if not .AvailableSides }}none{{else}}{{.AvailableSides}}{{end}}
  Available toppings: {{ if not .AvailableToppings }}none{{else}}{{.AvailableToppings}}{{end}}
`
