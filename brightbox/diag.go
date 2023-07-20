package brightbox

import (
	"errors"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"golang.org/x/oauth2"
)

func brightboxFromErrSlice(err error) diag.Diagnostics {
	return diag.Diagnostics{
		brightboxFromErr(err),
	}
}

func brightboxFromErr(err error) diag.Diagnostic {
	var brightboxError *brightbox.APIError
	if errors.As(err, &brightboxError) {
		log.Printf("[DEBUG] Returning API Diagnostic")
		return apiErrorDiagnostic(brightboxError)
	}
	var oauthError *oauth2.RetrieveError
	if errors.As(err, &oauthError) {
		log.Printf("[DEBUG] Returning Oauth2 Diagnostic")
		return oauth2ErrorDiagnostic(oauthError)
	}
	log.Printf("[DEBUG] Returning Normal Error with type %T", err)
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  err.Error(),
	}
}

func oauth2ErrorDiagnostic(err *oauth2.RetrieveError) diag.Diagnostic {
	descSlice := []string{}
	if err.ErrorDescription != "" {
		descSlice = append(descSlice, err.ErrorDescription)
	}
	if err.ErrorURI != "" {
		descSlice = append(descSlice, err.ErrorURI)
	}
	result := diag.Diagnostic{
		Severity: diag.Error,
		Summary:  err.ErrorCode,
		Detail:   strings.Join(descSlice, ", "),
	}
	return result
}

func apiErrorDiagnostic(err *brightbox.APIError) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  err.ErrorName,
		Detail:   strings.Join(err.Errors, ", "),
	}
}
