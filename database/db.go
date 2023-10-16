package db

import (
	"fmt"
	"strings"

	bbolt "github.com/PoulIgorson/sub_engine_fiber/database/bbolt"
	. "github.com/PoulIgorson/sub_engine_fiber/database/errors"
	pocketbase "github.com/PoulIgorson/sub_engine_fiber/database/pocketbase"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

func OpenBbolt(path string) (*bbolt.DataBase, error) {
	db, err := bbolt.Open(path)
	return db, err
}

func OpenPocketBase(address, identity, password string, isAdmin bool, updateCollections ...bool) (*pocketbase.DataBase, error) {
	fmt.Println("remote")
	return pocketbase.Open(address, identity, password, isAdmin, updateCollections...), nil
}

// Error if not valid data to authenticate
func OpenPocketBaseLocal(email, password string, isAdmin ...bool) (*pocketbase.DataBase, error) {
	fmt.Println("local")
	isAdmin_ := false
	if len(isAdmin) > 0 {
		isAdmin_ = isAdmin[0]
	}
	app := pocketbase.NewLocal(email, password, isAdmin_)
	if app == nil {
		return nil, NewErrorf("pocketbase not opened")
	}
	if isAdmin_ {
		_, err := app.Filter("users", nil)
		if err != nil && strings.Contains(err.Error(), ".token:") {
			if strings.Contains(err.Error(), "refused") {
				return nil, err
			}
			status, respI, err := GetJSONResponse(
				"POST", app.Address()+"/api/admins",
				Headers{"Content-Type": "application/json"},
				Data{
					"email":           email,
					"password":        password,
					"passwordConfirm": password,
				},
			)
			if err != nil {
				if strings.Contains(err.Error(), "admin authorization token") {
					return nil, NewErrorf("pb: create new admin does not support")
				}
				return nil, NewErrorf("pb: createAdmin.error: %v", err)
			} else if status != 200 {
				return nil, NewErrorf("pb: create admin: %+v", respI)
			}
		}
	}
	return pocketbase.OpenWith(app), nil
}
