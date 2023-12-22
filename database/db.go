package db

import (
	bbolt "github.com/PoulIgorson/sub_engine_fiber/database/bbolt"
	pocketbase "github.com/PoulIgorson/sub_engine_fiber/database/pocketbase"
	pocketbaselocal "github.com/PoulIgorson/sub_engine_fiber/database/pocketbaselocal"
)

func OpenBbolt(path string) (*bbolt.DataBase, error) {
	return bbolt.Open(path)
}

func OpenPocketBase(address, identity, password string, isAdmin bool, updateCollections ...bool) (*pocketbase.DataBase, error) {
	return pocketbase.Open(address, identity, password, isAdmin, updateCollections...), nil
}

// Error if not valid data to authenticate
func OpenPocketBaseLocal() (*pocketbaselocal.DataBase, error) {
	return pocketbaselocal.New(), nil
}
