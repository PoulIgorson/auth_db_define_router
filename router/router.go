// Package router implements setting routing for site.
package router

import (
	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	urls "github.com/PoulIgorson/sub_engine_fiber/router/urls"
)

// Router setting handlers on url
func Router(app *fiber.App, db_ *db.DB) {
	for _, url := range urls.UrlPatterns {
		app.Add(url.Method, url.Path, url.Handler(db_, urls.UrlPatterns, urls.AdminPatterns))
	}
	for _, url := range urls.AdminPatterns {
		app.Add(url.Method, url.Path, url.Handler(db_, urls.UrlPatterns, urls.AdminPatterns))
	}
}
