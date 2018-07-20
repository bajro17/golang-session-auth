package main

import (
	"fmt"
	"net/http"
	"html/template"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB
var tpl *template.Template

type User struct {
	gorm.Model
	Email    string `sql:"email"`
	Username string `sql:"username"`
	Password string `sql:"password"`
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	db, _ = gorm.Open("mysql", "root:@/ecom?charset=utf8&parseTime=True&loc=Local")
	
  defer db.Close()
  r := mux.NewRouter()
  http.Handle("/", r)
  r.HandleFunc("/", index).Methods("GET")
  r.HandleFunc("/register", register).Methods("GET","POST")
  r.HandleFunc("/login", login).Methods("GET","POST")
  r.HandleFunc("/logout", logout).Methods("POST")
  r.NotFoundHandler = http.HandlerFunc(NotFound)
  
  
  http.ListenAndServe(":9000", nil)

}

var Store = sessions.NewCookieStore([]byte("secret-password"))

func NotFound(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w,"404")
}

func index(w http.ResponseWriter, r *http.Request) {
	
	fmt.Fprint(w,"test")
}

func register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tpl.ExecuteTemplate(w, "register.html", nil)
	case "POST":
	 {
		r.ParseForm()

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		email := r.Form.Get("email")

		

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
		user := &User{
			Email:    email,
			Username: username,
			Password: string(hash),
		}
		db.Create(user)

		http.Redirect(w, r, "/login/", 302)

	}
}
	
}

func login(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session")

	if err != nil {
		fmt.Println("error identifying session")
		tpl.Execute(w, nil)
		return
	}

	switch r.Method {
	case "GET": 
	m := ""
	if flashes := session.Flashes(); len(flashes) > 0 {
		session.Save(r,w)
		for _, msg := range flashes {
			m = msg.(string)
		}
	}
		tpl.ExecuteTemplate(w,"login.html", m)

	case "POST":

		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		u := User{}
		db.Where("username = ?", username).First(&u)
		
		hash := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		
		if username != "" && password != "" {
			
			if u.Username == username && hash == nil {
			session.Values["auth"] = "true"
			session.Values["username"] = username
			session.AddFlash("You login successful!")
			session.Save(r, w)
			
			
			http.Redirect(w, r, "/login", 302)
			return
			}
			session.AddFlash("Somethng wrong!")
			session.Save(r, w)
			
			
			http.Redirect(w, r, "/login", 302)
		}
		
		
		session.Save(r, w)
		http.Redirect(w, r, "/login/", 302)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session")
	if err == nil { //If there is no error, then remove session
		if session.Values["auth"] != "false" {
			session.Values["auth"] = "false"
			session.Save(r, w)
		}
	}
	http.Redirect(w, r, "/login", 302)
	//redirect to login irrespective of error or not
}

func IsLoggedIn(r *http.Request) bool {
	session, _ := Store.Get(r, "session")
	if session.Values["auth"] == "true" {
		return true
	}
	return false
}
