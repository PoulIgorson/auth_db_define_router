package db

import (
	. "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"

	bbolt "github.com/PoulIgorson/sub_engine_fiber/database/bbolt"
)

func OpenBbolt(path string) (DB, error) {
	db, err := bbolt.Open(path)
	return db, err
}
