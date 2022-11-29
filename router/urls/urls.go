// Package urls containing list of struct Url[method, path, handler, name] for routing
package urls

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	db "auth_db_define_router/datebase"
	admin "auth_db_define_router/router/admin"
	auth "auth_db_define_router/router/auth"
	functions "auth_db_define_router/router/functions"
)

type Url struct {
	Method      string
	Path        string
	Handler     func(*db.DB, ...interface{}) fiber.Handler
	Name        string
	DisplayName string
}

var UrlPatterns = []*Url{
	&Url{"Get", "/", functions.IndexPage, "index", ""},

	&Url{"All", "/login", auth.LoginPage, "auth-login", ""},
	&Url{"Get", "/logout", auth.APILogout, "auth-logout", ""},
	&Url{"Get", "/new_password", auth.APINewPassword, "auth-new-password", ""},
	&Url{"All", "/registration", auth.RegistrationPage, "auth-registration", ""},

	&Url{"Get", "/admin", admin.IndexPage, "admin", "Админ"},
}

var AdminPatterns = []*Url{
	&Url{"Get", "/admin", admin.IndexPage, "admin", "Админ"},
}

func GetUrl(name string) *Url {
	for _, url := range UrlPatterns {
		if url.Name == name {
			return url
		}
	}
	return nil
}

func GetUrlOfPath(path string) *Url {
	patt := strings.Split(path, "/")[1:]
	for _, url := range UrlPatterns {
		url_patt := strings.Split(url.Path, "/")[1:]
		if len(patt) != len(url_patt) {
			continue
		}
		count := 0
		if index := GetUrl("index"); false && path == index.Path {
			return index
		}
		for i := 0; i < len(patt); i++ {
			if patt[i] != url_patt[i] && (len(url_patt[i]) == 0 || url_patt[i][0] != ':') {
				break
			} else {
				count++
			}
		}
		if count == len(patt) {
			return url
		}
	}
	return nil
}
