package cache_test

import (
	"fmt"
	"log"

	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func ExampleDataBase() {
	// open the database
	db, err := cache.GetDB(tests.TempFile())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err = db.Put("key", []byte("some string of values")); err != nil {
		log.Fatal(err)
	}

	data, err := db.Get("key")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// Output:
	// some string of values

	if err = db.Destroy(); err != nil {
		panic(err)
	}
}
