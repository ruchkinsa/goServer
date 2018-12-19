package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"html/template"
	"log"
	"net"
	"net/http"
	"path"
	"time"
	//"github.com/go-chi/render"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	"../model"
)

type Config struct {
	PublicPath          string
	PublicPathJS        http.FileSystem
	PublicPathCSS       http.FileSystem
	PublicPathTemplates http.FileSystem
}

type page struct {
	Title string
	Body  template.HTML //[]byte
	Users []*model.User
}

const lenPath = len("/")

var templates = make(map[string]*template.Template)

type ClaimsJWT struct {
	Name string `json: "name"`
	jwt.StandardClaims
}

//https://stackoverflow.com/questions/36236109/go-and-jwt-simple-authentication
// cookie - https://github.com/Unknwon/build-web-application-with-golang_EN/blob/master/eBook/06.1.md
/*

c := &http.Cookie{Name: "jwt", Value: str}
    http.SetCookie(w, c)
    w.Header().Set("Location", "/foo")
    w.WriteHeader(http.StatusFound)
*/

/*
type StandardClaims struct {
    Audience  string `json:"aud,omitempty"`	 	имя клиента для которого токен выпущен.
    ExpiresAt int64  `json:"exp,omitempty"`		срок действия токена.
    Id        string `json:"jti,omitempty"`		уникальный идентификатор токен (нужен, чтобы нельзя был «выпустить» токен второй раз)
    IssuedAt  int64  `json:"iat,omitempty"`		время выдачи токена.
    Issuer    string `json:"iss,omitempty"` 	адрес или имя удостоверяющего центра.
    NotBefore int64  `json:"nbf,omitempty"`		время, начиная с которого может быть использован (не раньше чем).
    Subject   string `json:"sub,omitempty"`		идентификатор пользователя. Уникальный в рамках удостоверяющего центра, как минимум.
}


*/

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS512", []byte("secret"), nil)
	_, tokenString, _ := tokenAuth.Encode(jwt.MapClaims{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

func Start(cfg Config, m *model.Model, listener net.Listener) {

	r := chi.NewRouter()
	// routers:
	// routers: protected
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth)) // Seek, verify and validate JWT tokens
		r.Use(jwtauth.Authenticator)       // можно переопределить этот метод проверки

		r.Handle("/people", peopleHandler(m))
	})

	// routers: public
	r.Group(func(r chi.Router) {
		r.Handle("/people/{param}", peopleHandlerParam(m))
		r.Route("/login", func(r chi.Router) {
			r.Post("/", authHandler(m)) // POST
			r.Get("/", loginHandler)    // GET
		})
		r.Handle("/", indexHandler(m))
		// ways static data
		r.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(cfg.PublicPathCSS)))
		r.Handle("/js/*", http.StripPrefix("/js/", http.FileServer(cfg.PublicPathJS)))
		r.Handle("/templates/*", http.StripPrefix("/templates/", http.FileServer(cfg.PublicPathTemplates)))
		// назначаем обработчик, если запрошенный url не существует
		r.NotFound(error404Handler)
	})
	// templates: base
	templates["index"] = template.Must(template.ParseFiles(path.Join(cfg.PublicPath, "templates", "layout.html"), path.Join(cfg.PublicPath, "templates", "index.html")))
	templates["error"] = template.Must(template.ParseFiles(path.Join("web", "templates", "layout.html"), path.Join("web", "templates", "error.html")))
	// server: settings
	server := &http.Server{
		Handler:        r,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 16}
	// server: run
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
		log.Println("authHandler->user: ")
		log.Println(user)
		if user != nil {
			// создание и запись данных о пользователе в сессию/БД/cookie
			token, err := createTokenJWT(user)
			if err != nil {
				log.Println("Error creating JWT token: ", err)
				errorHandler(w, r, http.StatusInternalServerError)
				return
			}
			log.Println("JWT token: ", token)
			// cookie
			jwtCookie := &http.Cookie{}
			jwtCookie.Name = "jwt"
			jwtCookie.Value = token
			jwtCookie.Path = "/"
			jwtCookie.Expires = time.Now().Add(time.Hour * 12)
			http.SetCookie(w, jwtCookie)
			// redirect to URL
			http.Redirect(w, r, "/people", 301)
		}
		p := page{Title: "Login", Body: template.HTML("<b>User not found!<b>")}
		renderTemplate(w, r, "login", &p)

	})
}

func createTokenJWT(m *model.User) (string, error) {
	claims := ClaimsJWT{
		"testName",
		jwt.StandardClaims{
			Id:        m.Login,                              //уникальный идентификатор токен (нужен, чтобы нельзя было «выпустить» токен второй раз)
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(), //срок действия токена
			/*
				Audience  string `json:"aud,omitempty"`	 	имя клиента для которого токен выпущен.
				ExpiresAt int64  `json:"exp,omitempty"`		срок действия токена.
				Id        string `json:"jti,omitempty"`		уникальный идентификатор токен (нужен, чтобы нельзя был «выпустить» токен второй раз)
				IssuedAt  int64  `json:"iat,omitempty"`		время выдачи токена.
				Issuer    string `json:"iss,omitempty"` 	адрес или имя удостоверяющего центра.
				NotBefore int64  `json:"nbf,omitempty"`		время, начиная с которого может быть использован (не раньше чем).
				Subject   string `json:"sub,omitempty"`		идентификатор пользователя. Уникальный в рамках удостоверяющего центра, как минимум.
			*/
		},
	}

	/*token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenEncodeString, err := token.SignedString([]byte("secret"))
	*/

	_, tokenEncodeString, err := tokenAuth.Encode(claims)

	if err != nil {
		return "", err
	}
	return tokenEncodeString, nil
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
