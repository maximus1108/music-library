package artist

import (
	"encoding/json"
	"fmt"
	"go-api/pkg/driver"
	"io/ioutil"
	"net/http"
	"strings"

	aDriver "github.com/arangodb/go-driver"
)

//Artist stuff
type Artist struct {
	Key         string             `json:"_key,omitempty"`
	ID          aDriver.DocumentID `json:"_id,omitempty"`
	Rev         string             `json:"_rev,omitempty"`
	Name        string             `json:"name"`
	RealName    string             `json:"real_name"`
	Nationality string             `json:"nationality"`
	Role        string             `json:"role,omitempty"`
}

//NewRepo stuff
func NewRepo(db aDriver.Database) Repo {
	return ArangoRepo{
		db,
	}
}

//ArangoRepo stuff
type ArangoRepo struct {
	db aDriver.Database
}

//Create astuff
func (r ArangoRepo) Create(a Artist) (Artist, error) {

	artists, err := r.db.Collection(nil, "artists")

	if err != nil {
		fmt.Println("error finding artists collection", err)
		return a, err
	}

	meta, err := artists.CreateDocument(nil, a)

	if err != nil {
		fmt.Println("error creating artist", err)
		return a, err
	}

	fmt.Printf("Created document in collection '%s' in database '%s'\n", artists.Name(), r.db.Name())

	fmt.Println(meta)

	return a, nil
}

//Fetch stuff
func (r ArangoRepo) Fetch() ([]Artist, error) {

	var a Artist
	var artists []Artist
	query := "FOR d IN artists RETURN d"

	cursor, err := r.db.Query(nil, query, nil)

	if err != nil {
		fmt.Println("cannot get artists", err)
		return artists, err
	}

	defer cursor.Close()

	for cursor.HasMore() {

		_, err := cursor.ReadDocument(nil, &a)

		if err != nil {
			fmt.Println("cannot get artist", err)
			return artists, err
		}

		artists = append(artists, a)

	}

	fmt.Println(artists)

	return artists, nil

}

//Repo stuff
type Repo interface {
	Create(a Artist) (Artist, error)
	Fetch() ([]Artist, error)
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
func (a Handler) Create(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	var data Artist
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

	a.repo.Create(data)

}

//Fetch stuff
func (a Handler) Fetch(w http.ResponseWriter, r *http.Request) {

	artists, err := a.repo.Fetch()

	if err != nil {
		fmt.Println("error fetching artists", err)

	}

	json.NewEncoder(w).Encode(artists)
}
