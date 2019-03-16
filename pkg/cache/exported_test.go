package cache_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/harrybrwn/apizza/pkg/cache"
)

func ExampleDataBase() {
	// open the database
	db, err := cache.GetDB(tempfile())
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
}

func tempfile() string {
	// f, err := ioutil.TempFile("", "apizza-")
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
