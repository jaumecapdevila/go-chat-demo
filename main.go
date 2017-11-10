package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/jaumecapdevila/chat/auth"
	"github.com/jaumecapdevila/chat/config"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/signature"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func init() {
	config := config.Load()
	gomniauth.SetSecurityKey(signature.RandomKey(64))
	gomniauth.WithProviders(
		github.New(
			config.Oauth.Github.Key,
			config.Oauth.Github.Secret,
			config.Oauth.Github.Callback,
		),
	)
}

func main() {
	var port = flag.String("port", ":8080", "The port of the application.")
	flag.Parse()
	r := newRoom()
	//r.tracer = trace.New(os.Stdout)
	http.Handle("/", auth.Must(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	// we dont need an object because this function doesn't need to store any state
	http.HandleFunc("/auth/", auth.LoginHandler)
	http.Handle("/room", r)

	// Run the chat in other goroutine
	// in order to let the main goroutine to run the webserver
	go r.run()
	log.Println("Starting web server on", *port)
	// Start the web server
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
