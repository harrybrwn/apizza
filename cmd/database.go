package cmd

import (
	"apizza/dawg"
	"apizza/pkg/config"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

type dataBase struct {
	DB *bolt.DB
}

func (db *dataBase) onCheckExists(bucket, key string) error {
	return nil
}

// func (db *dataBase) checkexists(bucket, key string) bool {
// 	var exists bool
// 	err := db.DB.View(func(tx *bolt.Tx) error {
// 		// b := tx.Bucket([]byte(bucket))
// 		// rawval := b.Get(key)
// 		if rawval == nil {
// 			exists = false
// 			return nil
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		fmt.Println("Yes... you need to handle thins error in checkexists", err)
// 		return false
// 	}
// 	return exists
// }

func initDatabase() error {
	var err error
	dir := filepath.Join(config.ConfigFolder(), "cache")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModeDir)
	}
	path := filepath.Join(dir, "apizza.db")
	db, err = bolt.Open(path, 0600, nil)
	return err
}

func menuManagment() error {
	var (
		err          error
		menuIsCached = true
	)
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Menu"))
		return err
	})
	if err != nil {
		return err
	}

	err = db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte("Menu"))
		rawmenu := b.Get([]byte("menu"))
		if rawmenu == nil {
			menuIsCached = false
			menu, err = store.Menu()
			return err
		}
		menu = &dawg.Menu{}
		return json.Unmarshal(rawmenu, menu)
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if menuIsCached {
			return nil
		}
		b := tx.Bucket([]byte("Menu"))
		data, err := json.Marshal(menu)
		if err != nil {
			return err
		}
		return b.Put([]byte("menu"), data)
	})
	return err
}
