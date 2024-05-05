// Main - access gohm algorithm via web interface and go template
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/above-the-garage/spwnn"
)

var dict *spwnn.SpwnnDictionary

var addr = flag.String("addr", ":1025", "http service address put colon in front")

var templ = template.Must(
	template.New("gohm").
		Funcs(template.FuncMap{"gohm": gohm}).
		Parse(templateStr))

func initDict() {
	dict = spwnn.ReadDictionary(false)
}

func gohm(word string) []spwnn.SpwnnResult {
	results, _ := spwnn.CorrectSpelling(dict, word, false /* strictLen */)
	return results
}

// SpwnnPage uses a template to access the gohm algorithm
func SpwnnPage(w http.ResponseWriter, req *http.Request) {
	templ.Execute(w, req.FormValue("word"))
}

const templateStr = `
<html>

<head>
  <title>Spwnn associative memory spelling corrector - Stephen Clarke-Willson</title>
  <style>
    body {
      transform: scale(3);
      transform-origin: top left;
    }
  </style>
</head>

<body>
  <form action="/" name=form method="GET">
    <input maxLength=30 size=30 name=word value="" title="Word" autofocus>
    <input type=submit value="Submit" name=cmd>
  </form>
  <p>
  <br>
  {{if .}}
  Checking word: '{{.}}'
  <br><br>
  Associative Results:<br>
	{{$val := (. | gohm)}}
	{{range $val}}
	  {{.Word}} <br>
	{{end}}
	<br>
  {{end}}
  <br>
  </p>
</body>

</html>
`

func main() {
	flag.Parse()
	initDict()
	http.Handle("/", http.HandlerFunc(SpwnnPage))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
