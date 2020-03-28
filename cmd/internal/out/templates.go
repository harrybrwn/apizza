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
  products:{{ range .Products }}
    {{.Name}}
      code:     {{.Code}}
      options:{{ range $k, $v := .ReadableOptions }}
         {{$k}}: {{$v}}{{else}}None{{end}}
      quantity: {{.Qty}}{{end}}
  storeID: {{.StoreID}}
  method:  {{.ServiceMethod}}
  address: {{.Addr -}}
{{ if .Price }}
  price:   ${{ .Price -}}
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
