package brightbox

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/gophercloud/gophercloud"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	maxPort = 65535
	minPort = 1
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

func userDataHashSum(userData string) string {
	// Check whether the userData is not Base64 encoded.
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it
	v, base64DecodeError := base64Decode(userData)
	if base64DecodeError != nil {
		v = userData
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

func setPrimaryCloudIP(d *schema.ResourceData, cloudIP *brightbox.CloudIP) {
	d.Set("ipv4_address", cloudIP.PublicIP)
	d.Set("public_hostname", cloudIP.Fqdn)
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

func mustBeBase64Encoded(v interface{}, name string) ([]string, []error) {
	return stringValidateFunc(
		v,
		name,
		func(value string) bool { return !isBase64Encoded(value) },
		"%q must be base64-encoded",
	)
}

// ValidateCronString checks if the string is a valid cron layout
func ValidateCronString(v interface{}, name string) (warns []string, errors []error) {
	if _, err := cronexpr.Parse(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q: %s", name, err))
	}
	return
}

func http1Keys(v interface{}, name string) (warns []string, errors []error) {
	mapValue, ok := v.(map[string]interface{})
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be a Map", name))
		return
	}
	for k := range mapValue {
		if !validToken(k) {
			errors = append(errors, fmt.Errorf("Metadata key %s is an invalid token. Should be all lower case, with no underscores", k))

		}
	}
	return
}

// Token has to be lower case
func validToken(tok string) bool {
	for i := 0; i < len(tok); i++ {
		if !validHeaderFieldByte(tok[i]) {
			return false
		}
	}
	return true
}

// validHeaderFieldByte reports whether b is a valid byte in a header
// field name. RFC 7230 says:
//   header-field   = field-name ":" OWS field-value OWS
//   field-name     = token
//   tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
//           "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
//   token = 1*tchar
//
// Underscore isn't valid. Needs to be a hyphen as Swift silently
// converts otherwise.
func validHeaderFieldByte(b byte) bool {
	return int(b) < len(isTokenTable) && isTokenTable[b]
}

var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  false,
	'B':  false,
	'C':  false,
	'D':  false,
	'E':  false,
	'F':  false,
	'G':  false,
	'H':  false,
	'I':  false,
	'J':  false,
	'K':  false,
	'L':  false,
	'M':  false,
	'N':  false,
	'O':  false,
	'P':  false,
	'Q':  false,
	'R':  false,
	'S':  false,
	'T':  false,
	'U':  false,
	'W':  false,
	'V':  false,
	'X':  false,
	'Y':  false,
	'Z':  false,
	'^':  true,
	'_':  false,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}

func escapedString(attr interface{}) string {
	return url.PathEscape(attr.(string))
}

func escapedStringList(source []string) []string {
	dest := make([]string, len(source))
	for i, v := range source {
		dest[i] = escapedString(v)
	}
	return dest
}

func escapedStringMetadata(metadata interface{}) map[string]string {
	source := metadata.(map[string]interface{})
	dest := make(map[string]string, len(source))
	for k, v := range source {
		dest[strings.ToLower(k)] = url.PathEscape(v.(string))
	}
	return dest
}

func removedMetadataKeys(old interface{}, new interface{}) []string {
	oldMap := old.(map[string]interface{})
	newMap := new.(map[string]interface{})
	result := make([]string, 0, len(oldMap))
	for key := range oldMap {
		if newMap[key] == nil {
			result = append(result, strings.ToLower(key))
		}
	}
	return result
}

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	if _, ok := err.(gophercloud.ErrDefault404); ok {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s %s: %s", msg, d.Id(), err)
}

// getEnvVarWithDefault retrieves the value of the environment variable
// named by the key. If the variable is not present, return the default
//value instead.
func getenvWithDefault(key string, defaultValue string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultValue
}

// set the lock state of a resource based upon a boolean
func setLockState(client *brightbox.Client, isLocked bool, resource interface{}) error {
	if isLocked {
		return client.LockResource(resource)
	}
	return client.UnLockResource(resource)
}
