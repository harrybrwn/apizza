package cmdtest

import (
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
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

// OrderName is the name of all testing orders created by the cmdtest package.
const OrderName = "cmdtest.TestingOrder"

// NewTestOrder creates an order for testing.
func NewTestOrder() *dawg.Order {
	o := &dawg.Order{
		StoreID:   "4336",
		Address:   dawg.StreetAddrFromAddress(TestAddress()),
		FirstName: "Jimmy",
		LastName:  "James",
		OrderName: OrderName,
	}
	o.Init()
	return o
}

// TestConfigjson data.
var TestConfigjson = `
{
	"name":"joe","email":"nojoe@mail.com",
	"address":{
		"street":"1600 Pennsylvania Ave NW",
		"cityName":"Washington DC",
		"state":"","zipcode":"20500"
	},
	"card":{"number":"","expiration":"","cvv":""},
	"service":"Carryout"
}`
