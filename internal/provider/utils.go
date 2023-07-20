package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func setEnvDefault(target *types.String, envVarName string, defaultValue string) {
	if !target.IsUnknown() && !target.IsNull() {
		return
	}
	v := os.Getenv(envVarName)
	if v != "" {
		*target = types.StringValue(v)
	} else if defaultValue != "" {
		*target = types.StringValue(defaultValue)
	}
}

func setEnv(target *types.String, envVarName string) {
	setEnvDefault(target, envVarName, "")
}

func addDefaultsToConfig(data BrightboxProviderModel) BrightboxProviderModel {
	setEnv(&data.Account, accountEnvVar)
	setEnvDefault(&data.APIClient, clientEnvVar, defaultClientID)
	setEnvDefault(&data.APISecret, clientSecretEnvVar, defaultClientSecret)
	setEnvDefault(&data.APIURL, apiURLEnvVar, defaultBaseURL)
	setEnvDefault(&data.OrbitURL, orbitURLEnvVar, defaultOrbitBaseURL)
	setEnv(&data.UserName, usernameEnvVar)
	setEnv(&data.password, passwordEnvVar)
	return data
}

func providerModelToAuthDetails(data BrightboxProviderModel) authdetails {
	return authdetails{
		APIClient: data.APIClient.String(),
		APISecret: data.APISecret.String(),
		UserName:  data.UserName.String(),
		password:  data.password.String(),
		Account:   data.Account.String(),
		APIURL:    data.APIURL.String(),
		OrbitURL:  data.OrbitURL.String(),
	}
}

func configureClient(ctx context.Context, data BrightboxProviderModel) (*CompositeClient, diag.Diagnostics) {
	tflog.Debug(ctx, "Configuring Brightbox Clients")

	authd := providerModelToAuthDetails(data)

	apiclient, orbitclient, diags := authenticatedClients(ctx, authd)

	if apiclient != nil {
		tflog.Info(ctx, fmt.Sprintf("Brightbox Client configured for URL: %s", apiclient.ResourceBaseURL()))
	}
	if orbitclient != nil {
		tflog.Info(ctx, fmt.Sprintf("Orbit Client configured for URL: %s", orbitclient.ResourceBaseURL()))
	}

	composite := &CompositeClient{
		APIClient:   apiclient,
		OrbitClient: orbitclient,
	}

	return composite, diags
}

func validateConfig(ctx context.Context, data BrightboxProviderModel) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog.Debug(ctx, "Validating Config")
	if strings.HasPrefix(data.APIClient.ValueString(), appPrefix) {
		tflog.Debug(ctx, "Detected OAuth Application. Validating User details.")
		if data.UserName.IsUnknown() || data.UserName.IsNull() {
			diags.AddError("missing UserName",
				"The UserName is missing. Please supply a Username and One Time Authentication code")
		}
		if data.password.IsUnknown() || data.password.IsNull() {
			diags.AddError("missing Password",
				"The password is missing. Please supply a Username and One Time Authentication code")
		}
		if data.Account.IsUnknown() || data.Account.IsNull() {
			diags.AddError("missing Account", "Must specify Account with User Credentials")
		}
	} else {
		tflog.Debug(ctx, "Detected API Client.")
		if !(data.UserName.IsUnknown() || data.UserName.IsNull()) {
			diags.AddError("UserName found",
				"User Credentials should not be supplied with an API Client. To use User Credentials supply an 'app' client, not a 'cli' client.")
		}
		if !(data.password.IsUnknown() || data.password.IsNull()) {
			diags.AddError("UserName found",
				"User Credentials should not be supplied with an API Client. To use User Credentials supply an 'app' client, not a 'cli' client.")
		}
	}
	return diags
}
