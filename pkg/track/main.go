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
	Create(a Track) (Track, []error)
	Fetch() ([]Track, error)
}

//ArangoRepo stuff
type ArangoRepo struct {
	db aDriver.Database
}

//Create astuff
func (r ArangoRepo) Create(t Track) (Track, []error) {

	tracks, err := r.db.Collection(nil, "tracks")

	if err != nil {
		fmt.Println("error finding artists collection", err)
		return t, []error{err}
	}

	artists := t.Artists
	t.Artists = nil

	meta, err := tracks.CreateDocument(nil, t)

	if err != nil {
		fmt.Println("error creating artist", err)
		return t, []error{err}
	}

	fmt.Printf("Created document in collection '%s' in database '%s'\n", tracks.Name(), r.db.Name())

	appearsInCol, err := r.db.Collection(nil, "appearsIn")

	if err != nil {
		fmt.Println("Could not get appearsIn edge collection", err)
		return t, []error{err}
	}

	var edgesCreated []string
	var edgeErrs []error

	for _, artist := range artists {

		edge := AppearsIn{
			From: artist.ID,
			To:   meta.ID,
			Role: artist.Role,
		}

		fmt.Println(edge, artist)

		edgeMeta, err := appearsInCol.CreateDocument(nil, edge)

		if err == nil {
			edgesCreated = append(edgesCreated, string(edgeMeta.Key))
		}

		if err != nil {

			fmt.Println("Could not create edge for", artist, err)

			edgeErrs = append(edgeErrs, err)

			if _, _, err := appearsInCol.RemoveDocuments(nil, edgesCreated); err != nil {
				edgeErrs = append(edgeErrs, err)
			}

			if _, err = tracks.RemoveDocument(nil, string(meta.Key)); err != nil {
				edgeErrs = append(edgeErrs, err)
			}

			return t, edgeErrs
		}

		fmt.Printf("Edge document in collection '%s' in database '%s'\n", appearsInCol.Name(), r.db.Name())

	}

	return t, nil

}

//Fetch astuff
func (r ArangoRepo) Fetch() ([]Track, error) {

	var t Track
	var tracks []Track

	query := `
		FOR track IN tracks
			LET artistsByTrack=(
				FOR artist, appears IN ANY track appearsIn
				RETURN {
					name: artist.name,
					real_name: artist.real_name,
					nationality: artist.nationality,
					role: appears.role
				}
			)
			RETURN {
				title: track.title,
				artists: artistsByTrack 
			}
	`

	cursor, err := r.db.Query(nil, query, nil)
	defer cursor.Close()

	if err != nil {
		fmt.Println("cannot get tracks", err)
		return nil, err
	}

	for cursor.HasMore() {

		if _, err := cursor.ReadDocument(nil, &t); err != nil {
			fmt.Println("cannot get artist", err)
			return nil, err
		}

		fmt.Println(t)

		tracks = append(tracks, t)

	}

	return tracks, nil
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

	_, errs := t.repo.Create(data)

	json.NewEncoder(w).Encode(errs)

}

//Fetch stuff
func (t Handler) Fetch(w http.ResponseWriter, r *http.Request) {

	tracks, err := t.repo.Fetch()

	if err != nil {
		fmt.Println("Error getting tracks", err)
	}

	fmt.Println(tracks)

	json.NewEncoder(w).Encode(tracks)

}
