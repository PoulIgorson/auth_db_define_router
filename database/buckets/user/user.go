// Package user implements model of bucket.
package user

import (
	"encoding/json"

	db "github.com/PoulIgorson/auth_db_define_router/database"
)

// Role implements access to module site
type Role struct {
	Name   string `json:"name"`
	Access uint   `json:"access"`
}

var Admin = &Role{"admin", ^uint(0)}
var Guest = &Role{"guest", 0}

var Roles = []*Role{
	Guest, Admin,
}
var Redirects = map[*Role]string{
	Guest: "/",
	Admin: "/admin",
}

func SetRoles(roles []*Role) {
	Roles = append(Roles, roles...)
}

func SetRedirectsForRoles(redirects map[*Role]string) {
	for role, url := range redirects {
		Redirects[role] = url
	}
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

	ExtraFields map[string]any `json:"extraFields"`
}

// Save implements saving model in bucket.
func (this *User) Save(bucket *db.Bucket) error {
	return db.SaveBucket(bucket, this)
}

func Create(db_ *db.DB, userStr string) *User {
	if userStr == "" {
		return nil
	}
	users, _ := db_.Bucket("users")
	var user, ruser User

	json.Unmarshal([]byte(userStr), &user)
	ruserStr, err := users.GetOfField("login", user.Login)
	if err != nil {
		return nil
	}

	json.Unmarshal([]byte(ruserStr), &ruser)
	if user.Password != ruser.Password {
		return nil
	}
	ruser.Role = GetRole(ruser.Role.Name)

	return &ruser
}

func CheckUser(db_ *db.DB, userStr string) *User {
	return Create(db_, userStr)
}

func CheckUserBool(db_ *db.DB, userStr string) bool {
	return Create(db_, userStr) != nil
}
