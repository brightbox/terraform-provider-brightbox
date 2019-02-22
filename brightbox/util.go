package brightbox

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
)

func hash_string(
	v interface{},
) string {
	switch v.(type) {
	case string:
		return userDataHashSum(v.(string))
	default:
		return ""
	}
}

func userDataHashSum(user_data string) string {
	// Check whether the user_data is not Base64 encoded.
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it
	v, base64DecodeError := base64Decode(user_data)
	if base64DecodeError != nil {
		v = user_data
	}
	hash := sha1.Sum([]byte(v))
	return hex.EncodeToString(hash[:])
}

func assign_string(d *schema.ResourceData, target **string, index string) {
	if d.HasChange(index) {
		if *target == nil {
			var temp string
			*target = &temp
		}
		if attr, ok := d.GetOk(index); ok {
			**target = attr.(string)
		}
	}
}

func assign_string_set(d *schema.ResourceData, target *[]string, index string) {
	if d.HasChange(index) {
		*target = map_from_string_set(d, index)
	}
}

func map_from_string_set(d *schema.ResourceData, index string) []string {
	var temp []string
	if attr := d.Get(index).(*schema.Set); attr.Len() > 0 {
		temp = make([]string, attr.Len())
		for i, v := range attr.List() {
			temp[i] = v.(string)
		}
	}
	return temp
}

func flatten_string_slice(list []string) []interface{} {
	temp := make([]interface{}, len(list))
	for i, v := range list {
		temp[i] = v
	}
	return temp
}

func assign_int(d *schema.ResourceData, target **int, index string) {
	if d.HasChange(index) {
		var temp int
		if attr, ok := d.GetOk(index); ok {
			temp = attr.(int)
		}
		*target = &temp
	}
}

func assign_bool(d *schema.ResourceData, target **bool, index string) {
	if d.HasChange(index) {
		var temp bool
		if attr, ok := d.GetOk(index); ok {
			temp = attr.(bool)
		}
		*target = &temp
	}
}

func setPrimaryCloudIp(d *schema.ResourceData, cloud_ip *brightbox.CloudIP) {
	d.Set("ipv4_address", cloud_ip.PublicIP)
	d.SetPartial("ipv4_address")
	d.Set("public_hostname", cloud_ip.Fqdn)
	d.SetPartial("public_hostname")
}

// Base64Encode encodes data if the input isn't already encoded
// using base64.StdEncoding.EncodeToString. If the input is already base64
// encoded, return the original input unchanged.
func base64Encode(data string) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return data
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func isBase64Encoded(data string) bool {
	_, err := base64Decode(data)
	return err == nil
}

func base64Decode(data string) (string, error) {
	result, err := base64.StdEncoding.DecodeString(data)
	return string(result), err
}

func stringValidateFunc(v interface{}, name string, failureTest func(string) bool, formatString string) (warns []string, errors []error) {
	value := v.(string)
	if failureTest(value) {
		errors = append(errors, fmt.Errorf(formatString, name))
	}
	return
}

func mustNotBeEmptyString(v interface{}, name string) ([]string, []error) {
	return stringValidateFunc(
		v,
		name,
		func(value string) bool { return value == "" },
		"%q cannot be empty",
	)
}

func mustBeBase64Encoded(v interface{}, name string) ([]string, []error) {
	return stringValidateFunc(
		v,
		name,
		func(value string) bool { return !isBase64Encoded(value) },
		"%q must be base64-encoded",
	)
}
