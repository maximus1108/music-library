package driver

import (
	"github.com/arangodb/go-driver"
	arango "github.com/arangodb/go-driver/http"
)

//DB words
type DB struct {
	Arango driver.Database
}

//ConnectArango wordws
func ConnectArango(host string, dbName string, uname string, pword string) (DB, error) {

	db := DB{}

	conn, err := arango.NewConnection(arango.ConnectionConfig{
		Endpoints: []string{host},
	})

	if err != nil {
		return db, err
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(uname, pword),
	})

	if err != nil {
		return db, err
	}

	arango, err := c.Database(nil, dbName)

	if err != nil {
		return db, err
	}

	db.Arango = arango

	return db, nil

}
