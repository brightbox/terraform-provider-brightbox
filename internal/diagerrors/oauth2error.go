package diagerrors

import (
	"fmt"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/exp/slices"
)

// Oauth2ErrorDiagnostic wraps a brightbox API error from the authentication subsystem
type Oauth2ErrorDiagnostic brightbox.APIError

// Severity is always an error
func (e *Oauth2ErrorDiagnostic) Severity() diag.Severity {
	return diag.SeverityError
}

// Summary returns the AuthError field of the underlying error
func (e *Oauth2ErrorDiagnostic) Summary() string {
	return e.AuthError
}

// Detail returns the AuthErrorDescription of the underlying error
func (e *Oauth2ErrorDiagnostic) Detail() string {
	return e.AuthErrorDescription
}

// Equal compares the two diagnostics
func (e *Oauth2ErrorDiagnostic) Equal(other diag.Diagnostic) bool {
	o, ok := other.(*Oauth2ErrorDiagnostic)

	if !ok {
		return false
	}
	if e.StatusCode != o.StatusCode {
		return false
	}
	if e.Status != o.Status {
		return false
	}
	if e.AuthError != o.AuthError {
		return false
	}
	if e.AuthErrorDescription != o.AuthErrorDescription {
		return false
	}
	if e.ErrorName != o.ErrorName {
		return false
	}
	if !slices.Equal(e.Errors, o.Errors) {
		return false
	}
	if e.ParseError != o.ParseError {
		return false
	}
	return true
}

// Error implements the error interface
func (e *Oauth2ErrorDiagnostic) Error() string {
	return fmt.Sprintf("%s, %s", e.Summary(), e.Detail())
}
