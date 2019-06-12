package track

import (
	"encoding/json"
	"fmt"
	"go-api/pkg/artist"
	"go-api/pkg/driver"
	"io/ioutil"
	"net/http"
	"strings"

	aDriver "github.com/arangodb/go-driver"
)

//Track stuff
type Track struct {
	Key     string             `json:"_key,omitempty"`
	ID      aDriver.DocumentID `json:"_id,omitempty"`
	Rev     string             `json:"_rev,omitempty"`
	Title   string             `json:"title"`
	Artists []artist.Artist    `json:"artists,omitempty"`
}

//AppearsIn Info
type AppearsIn struct {
	From aDriver.DocumentID `json:"_from,omitempty"`
	To   aDriver.DocumentID `json:"_to,omitempty"`
	Role string             `json:"role"`
}

//NewRepo stuff
func NewRepo(db aDriver.Database) Repo {
	return ArangoRepo{
		db,
	}
}

//Repo stuff
type Repo interface {
	Create(a Track) (Track, error)
	// Fetch() ([]Track, error)
}

//ArangoRepo stuff
type ArangoRepo struct {
	db aDriver.Database
}

//Create astuff
func (r ArangoRepo) Create(t Track) (Track, error) {

	tracks, err := r.db.Collection(nil, "tracks")

	if err != nil {
		fmt.Println("error finding artists collection", err)
		return t, err
	}

	artists := t.Artists
	t.Artists = nil

	meta, err := tracks.CreateDocument(nil, t)

	if err != nil {
		fmt.Println("error creating artist", err)
		return t, err
	}

	fmt.Printf("Created document in collection '%s' in database '%s'\n", tracks.Name(), r.db.Name())

	appearsInCollection, err := r.db.Collection(nil, "appearsIn")

	if err != nil {
		fmt.Println("Could not get appearsIn edge collection", err)
		return t, err
	}

	for _, artist := range artists {

		edge := AppearsIn{
			From: artist.ID,
			To:   meta.ID,
			Role: artist.Role,
		}
		fmt.Println(edge, artist)

		if _, err = appearsInCollection.CreateDocument(nil, edge); err != nil {
			fmt.Println("Could not create edge", err)
			return t, err
		}

		fmt.Printf("Edge document in collection '%s' in database '%s'\n", appearsInCollection.Name(), r.db.Name())

	}

	return t, nil

}

//NewHandler stuff
func NewHandler(db driver.DB) Handler {
	return Handler{
		repo: NewRepo(db.Arango),
	}
}

//Handler stuff
type Handler struct {
	repo Repo
}

//Create stuff
func (t Handler) Create(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	var data Track
	if err = json.Unmarshal(b, &data); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	key := strings.Trim(data.Title, " ")
	key = strings.ToLower(key)
	key = strings.Replace(key, " ", "-", -1)

	data.Key = key

	t.repo.Create(data)

}
