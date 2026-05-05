package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// StructValidator validates requirement of struct fields
func StructValidator(s interface{}) error {
	err := validate.Struct(s)
	var errorArr []string
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errorArr = append(
				errorArr,
				fmt.Sprintf("%v field doesn't satisfy the %v constraint", err.Field(), err.Tag()),
			)
		}
	}

	if len(errorArr) > 0 {
		return errors.New(strings.Join(errorArr, ";\n"))
	} else {
		return err
	}
}
