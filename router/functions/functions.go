// Package functions implements handlers for pages.
package functions

import (
	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	user "github.com/PoulIgorson/sub_engine_fiber/database/buckets/user"
	"github.com/PoulIgorson/sub_engine_fiber/types"
)

// IndexPage returns handler for index page.
func IndexPage(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		context := fiber.Map{
			"pagename": "Главная",
			"menu":     urls[0],
		}
		cuserI := c.Context().UserValue("user")
		var cuser *user.User
		if cuserI != nil {
			cuser = cuserI.(*user.User)
		}
		context["user"] = cuser
		if c.Method() == "GET" && cuser != nil {
			context["notifies"] = types.Notifies(cuser.ID, true)
		}
		return c.Render("index", context)
	}
}
