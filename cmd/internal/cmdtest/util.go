package cmdtest

import (
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/tests"
)

// TestAddress returns a testing address
func TestAddress() *obj.Address {
	return &obj.Address{
		Street:   "1600 Pennsylvania Ave NW",
		CityName: "Washington",
		State:    "DC",
		Zipcode:  "20500",
	}
}

// TempDB will create a new database in the temp folder.
func TempDB() *cache.DataBase {
	db, err := cache.GetDB(tests.NamedTempFile("cmdtest", "apizza_tmp.db"))
	if err != nil {
		panic(err)
	}
	return db
}
