package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var validationErrors validator.ValidationErrors
		if ok := errors.As(err, &validationErrors); ok {
			var errorMsgs []string
			for _, e := range validationErrors {
				errorMsgs = append(errorMsgs, fmt.Sprintf("field '%s' failed on '%s' tag", e.Field(), e.Tag()))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errorMsgs, ", "))
		}
		return err
	}
	return nil
}
