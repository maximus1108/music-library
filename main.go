package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/arangodb/go-driver"
	arango "github.com/arangodb/go-driver/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type artist struct {
	Key         string            `json:"_key,omitempty"`
	ID          driver.DocumentID `json:"_id,omitempty"`
	Rev         string            `json:"_rev,omitempty"`
	Name        string            `json:"name"`
	RealName    string            `json:"real_name"`
	Nationality string            `json:"nationality"`
	Role        string            `json:"role"`
}

type routeHandler func(http.ResponseWriter, *http.Request)

type routeCreater func(driver.Database) routeHandler

type track struct {
	Key     string            `json:"_key,omitempty"`
	ID      driver.DocumentID `json:"_id,omitempty"`
	Rev     string            `json:"_rev,omitempty"`
	Title   string            `json:"title"`
	Artists []artist          `json:"artists,omitempty"`
}

type response struct {
	Message string `json:"message"`
}

type appearsIn struct {
	From driver.DocumentID `json:"_from,omitempty"`
	To   driver.DocumentID `json:"_to,omitempty"`
	Role string            `json:"role"`
}

func createRoute(fn routeCreater, db driver.Database) routeHandler {
	return fn(db)
}

func createArtist(db driver.Database) routeHandler {

	return func(w http.ResponseWriter, r *http.Request) {

		b, err := ioutil.ReadAll(r.Body)

		defer r.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		var data artist
		err = json.Unmarshal(b, &data)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		key := strings.Trim(data.Name, " ")
		key = strings.ToLower(key)
		key = strings.Replace(key, " ", "-", -1)

		data.Key = key

		artists, err := db.Collection(nil, "artists")

		if err != nil {
			fmt.Println("error finding artists collection", err)
			return
		}

		meta, err := artists.CreateDocument(nil, data)

		if err != nil {
			fmt.Println("error creating artist", err)
			return
		}

		fmt.Printf("Created document in collection '%s' in database '%s'\n", artists.Name(), db.Name())

		fmt.Println(meta)

	}

}

func createTrack(db driver.Database) routeHandler {

	return func(w http.ResponseWriter, r *http.Request) {

		b, err := ioutil.ReadAll(r.Body)

		defer r.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		var data track
		err = json.Unmarshal(b, &data)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		key := strings.Trim(data.Title, " ")
		key = strings.ToLower(key)
		key = strings.Replace(key, " ", "-", -1)

		data.Key = key

		tracks, err := db.Collection(nil, "tracks")

		if err != nil {
			fmt.Println("error finding artists collection", err)
			return
		}

		artists := data.Artists
		data.Artists = nil

		meta, err := tracks.CreateDocument(nil, data)

		if err != nil {
			fmt.Println("error creating artist", err)
			return
		}

		fmt.Printf("Created document in collection '%s' in database '%s'\n", tracks.Name(), db.Name())

		appearsInCollection, err := db.Collection(nil, "appearsIn")

		for _, artist := range artists {
			var edge appearsIn

			edge.From = artist.ID
			edge.To = meta.ID
			edge.Role = artist.Role

			_, err = appearsInCollection.CreateDocument(nil, edge)

			if err != nil {
				fmt.Println("Could not create edge", err)
				return
			}

			fmt.Printf("Edge document in collection '%s' in database '%s'\n", appearsInCollection.Name(), db.Name())

		}

	}

}

func getArtists(db driver.Database) routeHandler {

	return func(w http.ResponseWriter, r *http.Request) {

		var a artist
		var artists []artist
		query := "FOR d IN artists RETURN d"

		cursor, err := db.Query(nil, query, nil)

		if err != nil {
			fmt.Println("cannot get artists", err)
		}

		defer cursor.Close()

		for cursor.HasMore() {

			_, err := cursor.ReadDocument(nil, &a)

			if err != nil {
				fmt.Println("cannot get artist", err)
				return
			}

			artists = append(artists, a)

		}

		fmt.Println(artists)

		json.NewEncoder(w).Encode(artists)

	}

}

func getTracks(db driver.Database) routeHandler {

	return func(w http.ResponseWriter, r *http.Request) {

		var t track
		var tracks []track

		query := `
			FOR t IN tracks
				LET artistsByTrack=(
					FOR artist, appears IN ANY t appearsIn
					RETURN {
						name: artist.name,
						real_name: artist.real_name,
						nationality: artist.nationality,
						role: appears.role
					}
				)
				RETURN {
					title:t.title,
					artists: artistsByTrack 
			}
		`

		cursor, err := db.Query(nil, query, nil)

		if err != nil {
			fmt.Println("cannot get artists", err)
		}

		defer cursor.Close()

		for cursor.HasMore() {

			_, err := cursor.ReadDocument(nil, &t)

			if err != nil {
				fmt.Println("cannot get artist", err)
				return
			}

			fmt.Println(t)

			tracks = append(tracks, t)

		}

		fmt.Println(tracks)

		json.NewEncoder(w).Encode(tracks)

	}

}

func databaseConnection(connectionStr string, databaseName string) (driver.Database, error) {

	conn, err := arango.NewConnection(arango.ConnectionConfig{
		Endpoints: []string{connectionStr},
	})

	if err != nil {
		return nil, err
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("root", "PKjylTWeyjd1AGUX"),
	})

	if err != nil {
		return nil, err
	}

	db, err := c.Database(nil, databaseName)

	if err != nil {
		return nil, err
	}

	return db, nil

}

func main() {

	connectionStr := "http://localhost:8528"
	databaseName := "music"

	db, err := databaseConnection(connectionStr, databaseName)

	if err != nil {
		fmt.Printf("unable to connect to '%s' on '%s' :: %s%", databaseName, connectionStr, err)
		// http.Error(w, err.Error(), 500)
	}

	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	r.HandleFunc("/artists", createRoute(getArtists, db)).
		Methods("GET")

	r.HandleFunc("/artists", createRoute(createArtist, db)).
		Methods("POST")

	r.HandleFunc("/tracks", createRoute(createTrack, db)).
		Methods("POST")

	r.HandleFunc("/tracks", createRoute(getTracks, db)).
		Methods("GET")

	http.Handle("/", r)

	acceptedHeaders := handlers.AllowedHeaders([]string{"Origin", "Accept", "X-Requested-With", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"})
	acceptedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	acceptedOrigins := handlers.AllowedOrigins([]string{"*"})

	http.ListenAndServe(":8083", handlers.CORS(acceptedHeaders, acceptedMethods, acceptedOrigins)(r))

}
