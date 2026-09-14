package validator

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"

	gpv "github.com/go-playground/validator/v10"
)

type Validate struct {
	*gpv.Validate
}

func New() *Validate {
	v := gpv.New(gpv.WithRequiredStructEnabled())

	registerCustomValidationRules(v)

	return &Validate{v}
}

func (v *Validate) Struct(s any) error {
	err := v.StructCtx(context.Background(), s)
	if err == nil {
		return nil
	}

	var vErrs gpv.ValidationErrors

	if !errors.As(err, &vErrs) {
		return err
	}

	return errs.ErrValidation{
		Message: getMessage(vErrs[0]),
	}
}

func getMessage(f gpv.FieldError) string {
	fName := formatFieldName(f.StructField())

	switch f.Tag() {
	case "required":
		return fmt.Sprintf(Required, fName)
	case "max":
		return fmt.Sprintf(Max, fName, f.Param())
	case "min":
		return fmt.Sprintf(Min, fName, f.Param())
	case "username":
		return fmt.Sprintf(Username, fName)
	case "path":
		return fmt.Sprintf(Path, fName)
	default:
		return fmt.Sprintf(Default, fName, f.Tag())
	}
}

// formatFieldName formats field's name
// from "SomeFieldName" to "Some field name".
func formatFieldName(name string) string {
	s := strcase.ToDelimited(name, ' ')
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func registerCustomValidationRules(v *gpv.Validate) {
	panicOnErr := func(err error) {
		if err != nil {
			panic(fmt.Sprintf("validator: failed to register validation rule: %s", err))
		}
	}

	//	Regexes
	var (
		username1Regex = regexp.MustCompile(`^[a-zA-Z0-9._]+$`)
		username2Regex = regexp.MustCompile(`[a-zA-Z]`)
		pathRegex      = regexp.MustCompile(`^/?[a-zA-Z0-9._-]+(?:/[a-zA-Z0-9._-]+)*/?$`)
	)

	//	Username
	panicOnErr(v.RegisterValidation("username", func(fl gpv.FieldLevel) bool {
		username := fl.Field().String()

		if !username1Regex.MatchString(username) {
			return false
		}

		return username2Regex.MatchString(username)
	}))

	//	Path
	panicOnErr(v.RegisterValidation("path", func(fl gpv.FieldLevel) bool {
		path := fl.Field().String()
		if path == "" {
			return true
		}

		if !pathRegex.MatchString(path) {
			return false
		}

		for _, segment := range strings.Split(path, "/") {
			if segment == "." || segment == ".." {
				return false
			}
		}

		return true
	}))
}
