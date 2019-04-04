package out

import (
	"io"
	"text/template"
)

func tmpl(w io.Writer, tmplt string, a interface{}) error {
	t := template.New("apizza")
	template.Must(t.Parse(tmplt))
	return t.Execute(w, a)
}

var defaultOrderTmpl = `{{ .OrderName }}
  products:{{ range .Products }}
    {{.Name}}
      code:     {{.Code}}
	  options:  {{ range $k, $v := .ReadableOptions }}
	  	  {{$k}}: {{$v}}{{else}}None{{end}}
      quantity: {{.Qty}}{{end}}
  storeID: {{.StoreID}}
  method:  {{.ServiceMethod}}
`

var cartOrderTmpl = `  {{ .OrderName }} - {{ range .Products }} {{.Code}}, {{end}}
`

var menuCategoryTmpl = ``

var variantTmpl = `{{ .Name }} {{ .Code }}
	price:   {{ .Price }}
	product: {{ .ProductCode }}
`

var shorthandVariantTmpl = `{{ .Name }} {{ .Code }}`
