// Package urls containing list of struct Url[method, path, handler, name] for routing
package urls

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	admin "github.com/PoulIgorson/sub_engine_fiber/router/admin"
	auth "github.com/PoulIgorson/sub_engine_fiber/router/auth"
	functions "github.com/PoulIgorson/sub_engine_fiber/router/functions"
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
	&Url{"All", "/logout", auth.APILogout, "api-auth-logout", ""},
	&Url{"POST", "/new_password", auth.APINewPassword, "api-auth-new-password", ""},
	&Url{"POST", "/registration", auth.APIRegistration, "api-auth-registration", ""},

	&Url{"Get", "/admin", admin.IndexPage, "admin", "Админ"},
}

var AdminPatterns = []*Url{
	&Url{"Get", "/admin", admin.IndexPage, "admin", "Админ"},
}

func AddUrlPatterns(up []*Url) {
	UrlPatterns = concat(up, UrlPatterns)
}

func AddAdminPatterns(ap []*Url) {
	AdminPatterns = concat(ap, AdminPatterns)
}

func concat(a, b []*Url) []*Url {
	for _, url := range b {
		t := len(a)
		for _, url2 := range a {
			if url2.Name == url.Name {
				continue
			}
			t--
		}
		if t == 0 {
			a = append(a, url)
		}
	}
	return a
}

func GetUrl(name string) *Url {
	for _, url := range concat(UrlPatterns, AdminPatterns) {
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
