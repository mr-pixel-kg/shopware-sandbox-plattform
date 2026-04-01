package http

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (v *CustomValidator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			details := make([]string, 0, len(validationErrs))
			for _, validationErr := range validationErrs {
				fieldName := validationErr.Field()
				if fieldName == "" {
					fieldName = validationErr.StructField()
				}
				fieldName = lowerFirst(fieldName)

				switch validationErr.Tag() {
				case "required":
					details = append(details, fmt.Sprintf("%s is required", fieldName))
				case "email":
					details = append(details, fmt.Sprintf("%s must be a valid email address", fieldName))
				case "min":
					details = append(details, fmt.Sprintf("%s must be at least %s characters long", fieldName, validationErr.Param()))
				case "oneof":
					details = append(details, fmt.Sprintf("%s must be one of: %s", fieldName, validationErr.Param()))
				case "gte":
					details = append(details, fmt.Sprintf("%s must be greater than or equal to %s", fieldName, validationErr.Param()))
				case "uuid4", "uuid":
					details = append(details, fmt.Sprintf("%s must be a valid UUID", fieldName))
				default:
					details = append(details, fmt.Sprintf("%s is invalid", fieldName))
				}
			}
			return fmt.Errorf("%s", strings.Join(details, "; "))
		}
		return err
	}

	return nil
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToLower(value[:1]) + value[1:]
}
