package http

import (
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/Nurlan270/cloud-storage-go/internal/core/config"
)

// BuildSessionCookieName returns session cookie's name
// using app's name (e.g. APP_NAME="My App" -> "my_app_session").
func BuildSessionCookieName(conf config.Config) string {
	str := strcase.ToSnake(conf.GetAppName())
	return strings.Trim(str, "_") + "_session"
}
