package main

import (
	"fmt"
	"net/http"

	"go-api/pkg/artist"
	"go-api/pkg/driver"
	"go-api/pkg/track"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	host := "http://localhost:8528"
	dbname := "music"
	uname := "root"
	pword := "u9tTW65sgiW5e2GM"

	db, err := driver.ConnectArango(host, dbname, uname, pword)

	if err != nil {
		fmt.Printf("unable to connect to '%s' on '%s' :: %s", dbname, uname, err)
		return
	}

	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	a := artist.NewHandler(db)

	r.HandleFunc("/artists", a.Fetch).
		Methods("GET")

	r.HandleFunc("/artists", a.Create).
		Methods("POST")

	t := track.NewHandler(db)

	r.HandleFunc("/tracks", t.Create).
		Methods("POST")

	r.HandleFunc("/tracks", t.Fetch).
		Methods("GET")

	http.Handle("/", r)

	headers := handlers.AllowedHeaders([]string{"Origin", "Accept", "X-Requested-With", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	http.ListenAndServe(":8083", handlers.CORS(headers, methods, origins)(r))

}
