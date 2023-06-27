// Package auth implements interface for auth.
package auth

import (
	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	user "github.com/PoulIgorson/sub_engine_fiber/database/buckets/user"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
)

var IgnoreUrls = []string{
	"/", "/login", "/logout", "/registration",
}

// New return handler for auth.
func New(db_ *db.DB, ignoreUrls ...[]string) fiber.Handler {
	if len(ignoreUrls) > 0 {
		IgnoreUrls = ignoreUrls[0]
	}
	return myNew(db_)
}

func myNew(db_ *db.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userStr := c.Cookies("userCookie")
		cuser := user.CreateIfExists(db_, userStr)
		c.Context().SetUserValue("user", cuser)
		if Contains(IgnoreUrls, c.Path()) || cuser != nil {
			return c.Next()
		}
		return c.Redirect("/login")
	}
}
