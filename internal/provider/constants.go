package provider

import (
	"github.com/brightbox/gobrightbox/v2/endpoint"
)

const (
	defaultClientID     = "app-dkmch"
	defaultClientSecret = "uogoelzgt0nwawb"
	clientEnvVar        = "BRIGHTBOX_CLIENT"
	clientSecretEnvVar  = "BRIGHTBOX_CLIENT_SECRET"
	usernameEnvVar      = "BRIGHTBOX_USER_NAME"
	passwordEnvVar      = "BRIGHTBOX_PASSWORD"
	accountEnvVar       = "BRIGHTBOX_ACCOUNT"
	apiURLEnvVar        = "BRIGHTBOX_API_URL"
	orbitURLEnvVar      = "BRIGHTBOX_ORBIT_URL"

	defaultTimeoutSeconds = 10
	appPrefix             = "app-"
	defaultBaseURL        = endpoint.DefaultBaseURL
	defaultOrbitBaseURL   = endpoint.DefaultOrbitBaseURL
)
