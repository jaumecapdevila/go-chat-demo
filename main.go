package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/jaumecapdevila/chat/auth"
	"github.com/jaumecapdevila/chat/config"
	log "github.com/sirupsen/logrus"
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
	if _, err := os.Stat(config.Logger.Dir); os.IsNotExist(err) {
		err := os.MkdirAll(config.Logger.Dir, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	if _, err := os.Stat(path.Join(config.Logger.Dir, "dev.log")); os.IsNotExist(err) {
		_, err := os.Create(path.Join(config.Logger.Dir, "dev.log"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	f, err := os.OpenFile(path.Join(config.Logger.Dir, "dev.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(f)
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
	// Start the web server
	if err := http.ListenAndServe(*port, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
