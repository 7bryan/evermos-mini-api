package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	playground "github.com/go-playground/validator/v10"

	apperrors "github.com/7bryan/evermos-mini-api/internal/platform/errors"
)

type Validator struct {
	v *playground.Validate
}

func New() *Validator {
	v := playground.New(playground.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		for _, tag := range []string{"json", "form"} {
			name := strings.SplitN(f.Tag.Get(tag), ",", 2)[0]
			if name != "" && name != "-" {
				return name
			}
		}
		return f.Name
	})
	return &Validator{v: v}
}

func (val *Validator) Struct(s any) error {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}
	var verrs playground.ValidationErrors
	if errors.As(err, &verrs) {
		details := make([]string, 0, len(verrs))
		for _, fe := range verrs {
			details = append(details, describe(fe))
		}
		return apperrors.Validation("validation failed", details...)
	}
	return apperrors.Internal(err)
}

func describe(fe playground.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return fmt.Sprintf("%s must be at least %s%s", field, fe.Param(), unit(fe))
	case "max":
		return fmt.Sprintf("%s must be at most %s%s", field, fe.Param(), unit(fe))
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return field + " is invalid"
	}
}

func unit(fe playground.FieldError) string {
	switch fe.Kind() {
	case reflect.String:
		return " characters"
	case reflect.Slice, reflect.Array, reflect.Map:
		return " items"
	default:
		return ""
	}
}
