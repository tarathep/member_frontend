package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/tarathep/member_frontend/api"
	"github.com/tarathep/member_frontend/assets"
	"github.com/tarathep/member_frontend/model"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

// TEMPLATES
var loginTpl *template.Template
var mainTpl *template.Template

// Config provides basic configuration
type Config struct {
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// HTMLServer represents the web service that serves up HTML
type HTMLServer struct {
	server *http.Server
	wg     sync.WaitGroup
}

type MainPage struct {
	Name       string
	Permission string
	Members    []model.Member
}

// MustAssetString returns the asset contents as a string (instead of a []byte).
func MustAssetString(name string) string {
	//GET templates.go is bianary html pages
	return string(assets.MustAsset(name))
}

// Render a template, or server error.
func render(w http.ResponseWriter, r *http.Request, tpl *template.Template, name string, data interface{}) {
	buf := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Write(buf.Bytes())
}

// ----------- PAGES HANDLER -----------

// LoginHandler renders the second view template
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	session, _ := store.Get(r, "cookie-name")
	fmt.Println(session.Values["authenticated"])

	var message bool

	if session.Values["authenticated"] == false {
		session.Values["authenticated"] = nil
		message = true
		session.Save(r, w)
	}
	data := map[string]interface{}{
		"Message": message,
	}

	render(w, r, loginTpl, "login", data)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = nil
	session.Values["role"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
	} else if r.Method != "POST" {

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	session, _ := store.Get(r, "cookie-name")

	username := r.PostFormValue("inputEmail")
	password := r.PostFormValue("inputPassword")

	//AS ROLE AND CHECK PERMISSION
	auth := api.Login(username, password)
	if auth.ID != "" {
		//PASS
		session.Values["authenticated"] = true
		session.Values["role"] = auth.Role
		session.Values["name"] = auth.Name

		session.Save(r, w)
		http.Redirect(w, r, "/main", http.StatusSeeOther)
	} else {
		//NOT PASS
		session.Values["authenticated"] = false
		session.Values["role"] = ""
		session.Values["name"] = ""

		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func MainHandler(w http.ResponseWriter, r *http.Request) {

	// ------------ CHECK AUTH
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// ------------ PROCESS PAGE
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := MainPage{
		Name:       session.Values["name"].(string),
		Permission: session.Values["role"].(string),
		Members:    api.GetMembers(),
	}
	buf := new(bytes.Buffer)
	if err := mainTpl.ExecuteTemplate(buf, "main", data); err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Write(buf.Bytes())

}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
	} else if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	id := r.FormValue("id")

	//DELETE FROM BACKEND DATABASE BY ID
	api.DeleteMembers(id)
	http.Redirect(w, r, "/main", http.StatusSeeOther)
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		log.Fatal("error parseForm")
	} else if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	member := model.Member{
		ID:        r.PostFormValue("modal_edit_id"),
		FirstName: r.PostFormValue("modal_edit_firstname"),
		LastName:  r.PostFormValue("modal_edit_lastname"),
		Role:      r.PostFormValue("modal_edit_role"),
		Email:     r.PostFormValue("modal_edit_email"),
		Password:  r.PostFormValue("modal_edit_password"),
	}
	api.EditMembers(member)
	http.Redirect(w, r, "/main", http.StatusSeeOther)
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		log.Fatal("error parseForm")
	} else if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	member := model.Member{
		ID:        r.PostFormValue("modal_add_id"),
		FirstName: r.PostFormValue("modal_add_firstname"),
		LastName:  r.PostFormValue("modal_add_lastname"),
		Role:      r.PostFormValue("modal_add_role"),
		Email:     r.PostFormValue("modal_add_email"),
		Password:  r.PostFormValue("modal_add_password"),
	}
	fmt.Println(member)

	api.AddMembers(member)
	http.Redirect(w, r, "/main", http.StatusSeeOther)

}

// Start launches the HTML Server
func Start(cfg Config) *HTMLServer {
	// Setup Context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Handlers
	router := mux.NewRouter()
	router.HandleFunc("/", LoginHandler).Methods("GET")
	router.HandleFunc("/", AuthHandler).Methods("POST")
	router.HandleFunc("/logout", LogoutHandler).Methods("GET")
	router.HandleFunc("/main", MainHandler).Methods("GET")
	router.HandleFunc("/delete", DeleteHandler).Methods("POST")
	router.HandleFunc("/edit", EditHandler).Methods("POST")
	router.HandleFunc("/add", AddHandler).Methods("POST")

	// Create the HTML Server
	htmlServer := HTMLServer{
		server: &http.Server{
			Addr:           cfg.Host,
			Handler:        router,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}

	// Add to the WaitGroup for the listener goroutine
	htmlServer.wg.Add(1)

	// Start the listener
	go func() {
		fmt.Printf("\nHTMLServer : Service started : Host=%v\n", cfg.Host)
		htmlServer.server.ListenAndServe()
		htmlServer.wg.Done()
	}()

	return &htmlServer
}

// Stop turns off the HTML Server
func (htmlServer *HTMLServer) Stop() error {
	// Create a context to attempt a graceful 5 second shutdown.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("\nHTMLServer : Service stopping\n")

	// Attempt the graceful shutdown by closing the listener
	// and completing all inflight requests
	if err := htmlServer.server.Shutdown(ctx); err != nil {
		// Looks like we timed out on the graceful shutdown. Force close.
		if err := htmlServer.server.Close(); err != nil {
			fmt.Printf("\nHTMLServer : Service stopping : Error=%v\n", err)
			return err
		}
	}

	// Wait for the listener to report that it is closed.
	htmlServer.wg.Wait()
	fmt.Printf("\nHTMLServer : Stopped\n")
	return nil
}

func init() {

	loginHTML := MustAssetString("templates/login.html")
	loginTpl = template.Must(template.New("login").Parse(loginHTML))

	mainHTML := MustAssetString("templates/main.html")
	mainTpl = template.Must(template.New("main").Parse(mainHTML))

}

func main() {
	serverCfg := Config{
		Host:         "localhost:5000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	htmlServer := Start(serverCfg)
	defer htmlServer.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("main : shutting down")
}
