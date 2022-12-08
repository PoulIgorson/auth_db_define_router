// Package functions implements handlers for admin pages.
package functions

import (
	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/auth_db_define_router/database"
	user "github.com/PoulIgorson/auth_db_define_router/database/buckets/user"
)

// IndexPage returns handler for admin index page.
func IndexPage(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cuser := user.CheckUser(db_, c.Cookies("userCookie"))
		context := fiber.Map{
			"pagename":   "Админ",
			"menu":       urls[0],
			"admin_menu": urls[1],
			"user":       cuser,
		}
		return c.Render("admin/index", context)
	}
}
