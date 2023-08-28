// Package auth implements interface for auth.
package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/sub_engine_fiber/database/interfaces"
	user "github.com/PoulIgorson/sub_engine_fiber/models/user"
)

var IgnoreUrls = []string{
	"/", "/login", "/logout", "/registration",
}

// New return handler for auth.
func New(db_ db.DB, ignoreUrls ...[]string) fiber.Handler {
	if len(ignoreUrls) > 0 {
		IgnoreUrls = append(IgnoreUrls, ignoreUrls[0]...)
	}
	return myNew(db_)
}

func myNew(db_ db.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userStr := c.Cookies("userCookie")
		cuser := user.CreateIfExists(db_, userStr)
		if cuser == nil {
			cuser = (*user.User)(nil)
		}
		c.Context().SetUserValue("user", cuser)
		if ContainsPath(IgnoreUrls, c.Path()) || cuser != nil {
			return c.Next()
		}
		return c.Redirect("/login")
	}
}

func ContainsPath(urls []string, path string) bool {
outer:
	for _, url := range IgnoreUrls {
		if url == path {
			return true
		}
		if len(url) == 0 || len(path) == 0 {
			continue
		}
		tokensUrl := strings.Split(url, "/")
		tokensPath := strings.Split(path, "/")
		for i := range tokensUrl {
			if i == len(tokensPath) {
				continue outer
			}
			if len(tokensUrl[i]) == 0 || tokensUrl[i][0] == ':' {
				continue
			}
			if tokensUrl[i] != tokensPath[i] {
				continue outer
			}
		}
		return true
	}
	return false
}
