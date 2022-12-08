// Package functions implements handlers for pages.
package functions

import (
	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/auth_db_define_router/database"
	user "github.com/PoulIgorson/auth_db_define_router/database/buckets/user"
)

// IndexPage returns handler for index page.
func IndexPage(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		context := fiber.Map{
			"pagename": "Главная",
			"menu":     urls[0],
		}
		context["user"] = user.CheckUser(db_, c.Cookies("userCookie"))
		return c.Render("index", context)
	}
}
