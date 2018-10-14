package api

import (
	"encoding/json"
	_ "fmt"
	"github.com/go-chi/chi"
	"html/template"
	"log"
	"net"
	"net/http"
	"path"
	"time"
	//"github.com/go-chi/render"

	"../model"
)

type Config struct {
	PublicPath          string
	PublicPathJS        http.FileSystem
	PublicPathCSS       http.FileSystem
	PublicPathTemplates http.FileSystem
}

var templates = make(map[string]*template.Template)

type page struct {
	Title string
	Body  template.HTML //[]byte
	Users []*model.User
}

const lenPath = len("/")

func Start(cfg Config, m *model.Model, listener net.Listener) {

	r := chi.NewRouter()
	// routers:
	r.Handle("/people", peopleHandler(m))
	r.Handle("/people/{param}", peopleHandlerParam(m))
	//r.Handle("/login", authHandler())
	r.Route("/login", func(r chi.Router) {
		r.Post("/", authHandler(m)) // POST
		r.Get("/", loginHandler)    // GET
	})
	r.Handle("/", indexHandler(m))
	// определяем пути к статическим данным
	r.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(cfg.PublicPathCSS)))
	r.Handle("/js/*", http.StripPrefix("/js/", http.FileServer(cfg.PublicPathJS)))
	r.Handle("/templates/*", http.StripPrefix("/templates/", http.FileServer(cfg.PublicPathTemplates)))
	// назначаем обработчик, если запрошенный url не существует
	r.NotFound(error404Handler)

	templates["index"] = template.Must(template.ParseFiles(path.Join(cfg.PublicPath, "templates", "layout.html"), path.Join(cfg.PublicPath, "templates", "index.html")))
	templates["error"] = template.Must(template.ParseFiles(path.Join("web", "templates", "layout.html"), path.Join("web", "templates", "error.html")))

	server := &http.Server{
		Handler:        r,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16}

	go server.Serve(listener)
}

func error404Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	log.Println("log: error404Handler")
	log.Println(r.URL.Path)
	err := templates["error"].ExecuteTemplate(w, "layout", map[string]interface{}{"Error": http.StatusText(404), "Status": 404})
	if err != nil {
		log.Println("log: error404Handler->error: " + err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	log.Println("log: errorHandler")
	err := templates["error"].ExecuteTemplate(w, "layout", map[string]interface{}{"Error": http.StatusText(status), "Status": status})
	if err != nil {
		log.Println("log: errorHandler->error: " + err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("loginHandler")
	p := page{Title: "Login", Body: template.HTML("<p>Login page</p>")}
	renderTemplate(w, r, "login", &p)
}

/*
func authHandler(w http.ResponseWriter, r *http.Request) {
		log.Println("authHandler")
		p := page{Title: "Login", Body: template.HTML("<p>Login page</p>")}
		renderTemplate(w, r, "login", &p)
}
*/

func authHandler(m *model.Model) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("authHandler")
		login := r.PostFormValue("username")
		password := r.PostFormValue("password")
		log.Println("authHandler->user: " + login)
		log.Println("authHandler->password: " + password)
		if login == "" || password == "" {
			//errorHandler(w, r, http.StatusBadRequest)
			return
		}
		user, err := m.CheckLoginUser(login, password) // запрос данных из БД
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest)
			return
		}
		log.Println("authHandler->user-len: ")
		log.Println(len(user))
		p := page{Title: "Login", Body: template.HTML("<b>User not found!<b>")}
		tmpl := "login"
		if len(user) > 0 {
			// создание и запись данных j пользователе в сессию/БД
			//
			people, err := m.People()
			if err != nil {
				errorHandler(w, r, http.StatusBadRequest)
				return
			}
			p = page{Title: "people", Users: people}
			tmpl = "people"
		}
		renderTemplate(w, r, tmpl, &p)

	})
}

func indexHandler(m *model.Model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("indexHandler")
		p := page{Title: "Home", Body: template.HTML("<p>Home page</p>")}
		//renderTemplate(w, r, "index", &p)
		renderTemplate(w, r, "index", &p)
	})
}

func peopleHandler(m *model.Model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		people, err := m.People()
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest)
			return
		}
		/*js, err := json.Marshal(people)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest)
			return
		}		*/
		//fmt.Fprintf(w, string(js))
		p := page{Title: "people", Users: people}
		renderTemplate(w, r, "people", &p)
	})
}

func peopleHandlerParam(m *model.Model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkTitleValidator(r.URL.Path[lenPath:]) {
			errorHandler(w, r, http.StatusBadRequest)
			return
		}
		people, err := m.People()
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest)
			return
		}

		js, err := json.Marshal(people)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest)
			return
		}
		//fmt.Fprintf(w, string(js))
		p := page{Title: "people2", Body: template.HTML(string(js))}
		renderTemplate(w, r, "people", &p)
	})
}
