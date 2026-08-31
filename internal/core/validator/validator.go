package validator

import (
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

	//	Custom username validation rule
	var re = regexp.MustCompile("^[a-zA-Z0-9._-]+$")

	err := v.RegisterValidation("username", func(fl gpv.FieldLevel) bool {
		username := fl.Field().String()

		return re.MatchString(username)
	})
	if err != nil {
		panic(fmt.Sprintf("validator: failed to register validation rule: %s", err))
	}

	return &Validate{v}
}

func (v *Validate) MapError(err error) error {
	var vErrors gpv.ValidationErrors

	if !errors.As(err, &vErrors) {
		return err
	}

	return errs.ErrValidation{
		Message: v.getMessage(vErrors[0]),
	}
}

func (v *Validate) getMessage(f gpv.FieldError) string {
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
	default:
		return fmt.Sprintf(Default, fName, f.Tag())
	}
}

// formatFieldName formats field's name
// from "SomeFieldName" to "Some field name".
func formatFieldName(n string) string {
	s := strcase.ToDelimited(n, ' ')
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
