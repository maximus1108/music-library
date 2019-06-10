package route

import (
	"net/http"

	"github.com/arangodb/go-driver"
)

type routeCreator func(driver.Database) http.HandlerFunc

//Route commented
type Route struct {
	DB      driver.Database
	Handler routeCreator
}

//CreateHandler commented
func (r Route) CreateHandler() http.HandlerFunc {
	return r.Handler(r.DB)
}

// func main() {
// 	testRoute := Route{}
// }

// func createArtist(db driver.Database) http.HandlerFunc {

// 	return func(w http.ResponseWriter, r *http.Request) {
