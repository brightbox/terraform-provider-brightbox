package diagerrors

import (
	"fmt"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"golang.org/x/exp/slices"
)

// APIErrorDiagnostic wraps a brightbox API error from the authentication subsystem
type APIErrorDiagnostic brightbox.APIError

// Severity is always an error
func (e *APIErrorDiagnostic) Severity() diag.Severity {
	return diag.SeverityError
}

// Summary returns the AuthError field of the underlying error
func (e *APIErrorDiagnostic) Summary() string {
	return e.ErrorName
}

// Detail returns the AuthErrorDescription of the underlying error
func (e *APIErrorDiagnostic) Detail() string {
	return strings.Join(e.Errors, ", ")
}

// Equal compares the two diagnostics
func (e *APIErrorDiagnostic) Equal(other diag.Diagnostic) bool {
	o, ok := other.(*APIErrorDiagnostic)

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
func (e *APIErrorDiagnostic) Error() string {
	return fmt.Sprintf("%s, %s", e.Summary(), e.Detail())
}
