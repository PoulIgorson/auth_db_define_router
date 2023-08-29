// Package user implements model of bucket.
package user

import (
	"encoding/json"

	db "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
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
	ID       any    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     *Role  `json:"role"`

	ExtraFields map[string]any `json:"extraFields"`
}

func (user User) Id() any {
	return user.ID
}

// Save implements saving model in bucket.
func (user *User) Save(bucket db.Table) error {
	return bucket.Save(user)
}

func Create(db_ db.DB, userStr string) *User {
	var user User
	json.Unmarshal([]byte(userStr), &user)
	d := map[string]any{}
	json.Unmarshal([]byte(userStr), &d)
	if roleI, ok := d["role"]; ok {
		role, ok := roleI.(map[string]any)
		if ok {
			if nameI, ok := role["name"]; ok {
				user.Role = GetRole(nameI.(string))
			}
		}
	}
	if _, ok := user.ID.(uint); !ok {
		if idF, ok := user.ID.(float64); ok {
			user.ID = uint(int(idF))
		}
	}
	return &user
}

func CreateIfExists(db_ db.DB, userStr string) *User {
	if !CheckUser(db_, userStr) {
		return nil
	}
	return Create(db_, userStr)
}

func (user User) Create(db_ db.DB, userStr string) db.Model {
	return Create(db_, userStr)
}

func CheckUser(db_ db.DB, userStr string) bool {
	user := Create(db_, userStr)
	if user.ID == "" {
		return false
	}
	userBct, _ := db_.Table("user", &User{})
	ruserM := userBct.Manager().Get(user.ID)
	if ruserM == nil {
		return false
	}
	if ruserM.(*User).Login != user.Login {
		return false
	}
	if ruserM.(*User).Password != user.Password {
		return false
	}
	return true
}

func (user User) Delete(db_ db.DB) error {
	userBct, _ := db_.Table("users", &User{})
	return userBct.Delete(user.ID)
}
