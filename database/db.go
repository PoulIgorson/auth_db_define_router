package db

import (
	pb "github.com/pocketbase/pocketbase"

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


func OpenPocketBaseLocal(app ...*pb.PocketBase) (*pocketbaselocal.DataBase, error) {
	return pocketbaselocal.New(app...), nil
}
