package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	//"strconv"
)

const lenPath = len("/")

type page struct {
	Title string
	Body  template.HTML //[]byte
}

var templates = make(map[string]*template.Template)
var titleValidator = regexp.MustCompile("^[a-zA-Z0-9/-]+$")

func init() {

}

func main() {
	// для отдачи сервером статичных файлов из папки public/static
	fs := http.FileServer(http.Dir("./public/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	templates["index"] = template.Must(template.ParseFiles(path.Join("templates", "layout.html"), path.Join("templates", "index.html")))
	templates["error"] = template.Must(template.ParseFiles(path.Join("templates", "layout.html"), path.Join("templates", "error.html")))

	// определяем порт для прослушивания
	port := flag.String("port", ":10000", "-port=:10000")
	flag.Parse()

	http.HandleFunc("/", makeHandler(handler))
	http.HandleFunc("/about/", makeHandler(aboutHandler))
	http.HandleFunc("/contacts", makeHandler(contactsHandler))
	http.HandleFunc("/profile", makeHandler(profileHandler))

	fmt.Printf("Started server on port %s", *port)
	log.Fatal(http.ListenAndServe(*port, nil))

}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if title == "" {
			title = "index"
		}
		log.Println("log: title=" + title)
		if !titleValidator.MatchString(title) {
			//http.NotFound(w, r)
			errorHandler(w, r, 404)
			return
		}
		fn(w, r, title)
	}
}

func handler(w http.ResponseWriter, r *http.Request, title string) {
	p, _ := loadPage(title)
	//renderTemplate(w, r, "index", p)
	renderTemplate(w, r, title, p)
}

func aboutHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := page{Title: title, Body: template.HTML("page about")}
	renderTemplate(w, r, "about", &p)
}

func contactsHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := page{Title: title, Body: template.HTML("page contacts")}
	renderTemplate(w, r, "contacts", &p)
}

func profileHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := page{Title: title, Body: template.HTML("page profile")}
	renderTemplate(w, r, "profile", &p)
}

func loadPage(title string) (*page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return &page{Title: title, Body: template.HTML("<p>Page not found</p>")}, nil
	}
	return &page{Title: title, Body: template.HTML(body)}, nil
}

func loadTemplates(nameTamplate string) (status int, err error) {
	if _, err := os.Stat(path.Join("templates", nameTamplate+".html")); os.IsNotExist(err) {
		// файл не существует
		return 404, err
	}
	templates[nameTamplate], err = template.New(nameTamplate).ParseFiles(path.Join("templates", "layout.html"), path.Join("templates", nameTamplate+".html"))
	return 200, err
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, p *page) {
	if _, err := templates[tmpl]; !err {
		if status, err := loadTemplates(tmpl); err != nil {
			log.Println(err.Error())
			errorHandler(w, r, status)
			return
		}
	}
	if err := templates[tmpl].ExecuteTemplate(w, "layout", p); err != nil {
		log.Println(err.Error())
		errorHandler(w, r, http.StatusInternalServerError)
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if err := templates["error"].ExecuteTemplate(w, "layout", map[string]interface{}{"Error": http.StatusText(status), "Status": status}); err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
}
