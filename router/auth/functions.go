// Package functions implements handlers for registration pages.
package functions

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"

	db "github.com/PoulIgorson/sub_engine_fiber/database"
	user "github.com/PoulIgorson/sub_engine_fiber/database/buckets/user"
	. "github.com/PoulIgorson/sub_engine_fiber/define"
	"github.com/PoulIgorson/sub_engine_fiber/types"
)

func dropPort(host string) string {
	index := strings.IndexRune(host, ':')
	if index != -1 {
		host = host[:index]
	}
	return host
}

// LoginPage returns handler for login page.
func LoginPage(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		context := fiber.Map{
			"pagename": "Login",
			"menu":     urls[0],
		}
		if c.Method() == "GET" {
			cuser := c.Context().UserValue("user").(*user.User)
			if cuser != nil {
				return c.Redirect("/")
			}
		} else if c.Method() == "POST" {
			var data map[string]string
			json.Unmarshal(c.Request().Body(), &data)

			users, _ := db_.Bucket("users", user.User{})
			cuserModel := users.Objects.Filter(db.Params{"Login": data["login"]}).First()
			if cuserModel == nil {
				return c.JSON(fiber.Map{"Status": "400", "login": "Логин не существует"})
			}

			cuser := cuserModel.(*user.User)
			if Hash([]byte(data["password"])) != cuser.Password {
				return c.JSON(fiber.Map{"Status": "400", "password": "Неверный пароль"})
			}

			strUser, _ := json.Marshal(cuser)
			cookie := fiber.Cookie{
				Name:        "userCookie",
				Value:       string(strUser),
				Path:        "/",
				Domain:      dropPort(c.Hostname()),
				Secure:      false,
				HTTPOnly:    false,
				SessionOnly: false,
			}
			c.Cookie(&cookie)

			url := "/"
			for role, curl := range user.Redirects {
				if cuser.Role == role {
					url = curl
					break
				}
			}
			types.NotifyInfo("Доброго времени суток, "+cuser.Login, cuser.ID)
			return c.JSON(fiber.Map{
				"Status":      "302",
				"redirectURL": url,
			})
		}
		return c.Render("registration/login", context)
	}
}

func APILogout(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cuser := c.Context().UserValue("user").(*user.User)
		c.ClearCookie("userCookie")
		types.NotifyInfo("Всего вам доброго, "+cuser.Login, cuser.ID)
		return c.Redirect("/")
	}
}

func APIRegistration(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data map[string]string
		json.Unmarshal(c.Request().Body(), &data)

		users, _ := db_.Bucket("users", user.User{})

		errors := map[string]string{}
		if len(data["login"]) < 4 {
			errors["login"] = "Слишком короткий логин"
		}

		if len(data["password1"]) < 8 {
			errors["password1"] = "Слишком короткий пароль"
		}

		if len(data["password2"]) < 8 {
			errors["password2"] = "Слишком короткий пароль"
		}

		if users.Objects.Filter(db.Params{"Login": data["login"]}).Count() > 0 {
			errors["login"] = "Логин существует"
		}

		if data["password1"] != data["password2"] {
			errors["password2"] = "Пароли не совпадают"
		}

		if len(errors) > 0 {
			errorsMap := fiber.Map{"Status": "400"}
			for field, err := range errors {
				errorsMap[field] = err
			}
			return c.JSON(errorsMap)
		}

		copyData := CopyMapAny(data)
		for _, k := range []string{"login", "password1", "password2", "role"} {
			delete(copyData, k)
		}

		cuser := user.User{
			Login:       data["login"],
			Password:    Hash([]byte(data["password1"])),
			Role:        user.GetRole("", ParseUint(data["role"])),
			ExtraFields: copyData,
		}
		cuser.Save(users)
		return c.JSON(fiber.Map{"Status": "302", "redirectURL": "/login"})
	}
}

func APINewPassword(db_ *db.DB, urls ...interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data map[string]string
		json.Unmarshal(c.Request().Body(), &data)

		password1 := data["password1"]
		password2 := data["password2"]

		if len(password1) < 8 {
			return c.JSON(fiber.Map{"Status": "500", "password1": "Слишком короткий пароль"})
		}
		if len(password2) < 8 {
			return c.JSON(fiber.Map{"Status": "500", "password2": "Слишком короткий пароль"})
		}
		if password1 != password2 {
			return c.JSON(fiber.Map{"Status": "500", "password1": "Пароли не совпадают"})
		}

		users, _ := db_.Bucket("users", user.User{})
		cuserModel := users.Objects.Filter(db.Params{"Login": data["login"]}).First()
		if cuserModel == nil {
			return c.JSON(fiber.Map{"Status": "500", "Error": "Логин не существует"})
		}
		cuser := cuserModel.(*user.User)
		cuser.Password = Hash([]byte(password1))
		cuser.Save(users)
		types.NotifyInfo("Пароль успешно изменен", cuser.ID)
		return c.JSON(fiber.Map{"Status": "200"})
	}
}
