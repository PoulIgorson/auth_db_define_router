// Package user implements model of bucket.
package user

import (
	"encoding/json"
	"fmt"

	db "auth_db_define_router/datebase"
	. "auth_db_define_router/define"
)

// Role implements access to module site
type Role struct {
	Name   string `json:"name"`
	Access uint   `json:"access"`
}

var Admin = &Role{"admin", 3}
var Checker = &Role{"checker", 2}
var Worker = &Role{"worker", 1}
var Guest = &Role{"guest", 0}

var Roles = []*Role{
	Admin, Checker, Worker, Guest,
}

func GetRole(name string, access ...uint) *Role {
	for _, role := range Roles {
		if role.Name == name || (len(access) > 0 && role.Access == access[0]) {
			return role
		}
	}
	return nil
}

// User presents model of bucket.
type User struct {
	ID       uint   `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     *Role  `json:"role"`
	Command  string `json:"command"`
}

// Save implements saving model in bucket.
func (this *User) Save(bucket *db.Bucket) error {
	// if object does not exists
	if _, err := bucket.Get(int(this.ID)); err != nil || this.ID == 0 {
		k, _ := bucket.Get(0)
		id := Atoi(k)
		if id == 0 {
			id++
		}
		this.ID = uint(id)
		bucket.Set(0, Itoa(id+1))
	}

	if this.Role.Name == "" {
		this.Role = Guest
	}
	buf, err := json.Marshal(this)
	if err != nil {
		fmt.Printf("Error of saving bucket `users`: %v", err.Error())
		return err
	}
	return bucket.Set(int(this.ID), string(buf))
}

func CheckUser(db_ *db.DB, userStr string) *User {
	if userStr == "" {
		return nil
	}
	users, err := db_.Bucket("users")
	if err != nil {
		return nil
	}
	var user User
	json.Unmarshal([]byte(userStr), &user)
	value, err := users.GetOfField("login", user.Login)
	if err != nil {
		return nil
	}
	var ruser User
	json.Unmarshal([]byte(value), &ruser)
	if user.Password != ruser.Password {
		return nil
	}
	ruser.Role = GetRole(ruser.Role.Name)
	return &ruser
}

func CheckUserBool(db_ *db.DB, userStr string) bool {
	return CheckUser(db_, userStr) != nil
}
