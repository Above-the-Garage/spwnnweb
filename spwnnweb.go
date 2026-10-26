// Main - access gohm algorithm via web interface and go template
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/above-the-garage/spwnn"
)

var dict *spwnn.SpwnnDictionary

var addr = flag.String("addr", ":8080", "http service address put colon in front")

var templ = template.Must(template.New("gohm").Parse(templateStr))

func initDict() {
	dict = spwnn.ReadDictionary("knownWords.txt")
}

// pageData is what the template shows
type pageData struct {
	Word    string
	Scoring string
	Length  string
	Results []spwnn.SpwnnResult
}

// SpwnnPage uses a template to access the gohm algorithm.
// Optional "?scoring=capped" (default) or "?scoring=original", and
// "?length=any" (default) or "?length=same"; "?strict" is kept as a
// shorthand for "?length=same"; "?help" shows the options.
func SpwnnPage(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	if query.Has("help") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, helpText)
		return
	}

	lengthMatch := spwnn.AnyLength
	if _, strict := query["strict"]; strict {
		lengthMatch = spwnn.SameLength
	}
	if query.Has("length") {
		var err error
		lengthMatch, err = parseLength(query.Get("length"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	scoring := spwnn.CappedPairs
	if query.Has("scoring") {
		var err error
		scoring, err = parseScoring(query.Get("scoring"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	data := pageData{
		Word:    req.FormValue("word"),
		Scoring: query.Get("scoring"),
		Length:  query.Get("length"),
	}
	if lengthMatch == spwnn.SameLength {
		data.Length = "same"
	}
	if data.Word != "" {
		data.Results, _ = spwnn.CorrectSpelling(dict, data.Word, lengthMatch, scoring)
	}
	templ.Execute(w, data)
}

const helpText = `spwnn spelling corrector

Query parameters:
  ?word=WORD          word to correct (required)
  &scoring=capped     credit each letter pair at most as often as the
                      dictionary word has it (default)
  &scoring=original   credit every repeated letter pair (old scoring)
  &length=any         keep best matches of any length (default)
  &length=same        keep only best matches the same length as WORD
  &strict             shorthand for &length=same
  ?help               show this help

Example: ?word=speling&length=same
`

// parseScoring converts the scoring query text to a spwnn.PairScoring
func parseScoring(name string) (spwnn.PairScoring, error) {
	switch name {
	case "capped":
		return spwnn.CappedPairs, nil
	case "original":
		return spwnn.OriginalPairs, nil
	}
	return spwnn.CappedPairs, fmt.Errorf("unknown scoring %q (want capped or original)", name)
}

// parseLength converts the length query text to a spwnn.LengthMatch
func parseLength(name string) (spwnn.LengthMatch, error) {
	switch name {
	case "any":
		return spwnn.AnyLength, nil
	case "same":
		return spwnn.SameLength, nil
	}
	return spwnn.AnyLength, fmt.Errorf("unknown length %q (want any or same)", name)
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
  <form action="/spwnn" name=form method="GET">
    <input maxLength=30 size=30 name=word value="" title="Word" autofocus>
    <input type=submit value="Submit" name=cmd>
    {{if .Scoring}}<input type=hidden name=scoring value="{{.Scoring}}">{{end}}
    {{if .Length}}<input type=hidden name=length value="{{.Length}}">{{end}}
  </form>
  <p>
  <br>
  {{if .Word}}
  Checking word: '{{.Word}}'
  <br><br>
  Associative Results:<br>
	{{range .Results}}
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
