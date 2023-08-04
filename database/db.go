package db

import (
	bbolt "github.com/PoulIgorson/sub_engine_fiber/database/bbolt"
	pocketbase "github.com/PoulIgorson/sub_engine_fiber/database/pocketbase"
)

func OpenBbolt(path string) (*bbolt.DataBase, error) {
	db, err := bbolt.Open(path)
	return db, err
}

func OpenPocketBase(address, identity, password string) (*pocketbase.DataBase, error) {
	return pocketbase.Open(address, identity, password), nil
}

func OpenPocketBaseLocal(identity, password string) (*pocketbase.DataBase, error) {
	app := pocketbase.NewLocal(identity, password)
	return pocketbase.OpenWith(app, map[string]*pocketbase.Bucket{}), nil
}
