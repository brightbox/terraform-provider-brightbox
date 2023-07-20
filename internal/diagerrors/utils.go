package diagerrors

import (
	"errors"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// FromErr wraps a Brightbox error as the relevant Terraform diagnostic
func FromErr(err error) diag.Diagnostic {
	var brightboxError *brightbox.APIError
	if errors.As(err, &brightboxError) {
		if brightboxError.AuthError != "" {
			result := Oauth2ErrorDiagnostic(*brightboxError)
			return &result
		}
		if brightboxError.ErrorName != "" {
			result := APIErrorDiagnostic(*brightboxError)
			return &result
		}
	}
	return diag.NewErrorDiagnostic(err.Error(), "")
}
