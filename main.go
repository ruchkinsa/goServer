package main

import (
	"html/template"
	"log"
	"net/http"
)

type Item struct {
	Caption string
	Link    string
}

type Links struct {
	CaptionActiveUrl string
	Items            [3]Item
}

const templatePage = `<h1>{{.CaptionActiveUrl}}</h1>
<u>
	{{range .Items}}
	<li><a href='{{.Link}}'>{{.Caption}}</a></li>
	{{end}}
</u>`

var report = template.Must(template.New("test").Parse(templatePage))
var content Links

func main() {
	content.Items[0].Caption = "Link1"
	content.Items[0].Link = "/link1"
	content.Items[1].Caption = "Link2"
	content.Items[1].Link = "/link2"
	content.Items[2].Caption = "Link3"
	content.Items[2].Link = "/link3"
	content.CaptionActiveUrl = "Home"

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:10000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	content.CaptionActiveUrl = "Home"
	for _, link := range content.Items {
		if r.URL.Path == link.Link {
			content.CaptionActiveUrl = link.Caption
		}
	}
	//if err := report.Execute(os.Stdout,  content);  err  !=  nil  {
	if err := report.Execute(w, content); err != nil {
		log.Fatal(err)
	}
	//fmt.Fprintf(w,  "URL.Path  = %q\n",  r.URL.Path)
}
