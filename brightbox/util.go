package brightbox

import (
	"crypto/sha1"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"os"
	"regexp"
	"sort"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/gophercloud/gophercloud"
	"github.com/gorhill/cronexpr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	serverRegexp           = regexp.MustCompile("^srv-.....$")
	serverGroupRegexp      = regexp.MustCompile("^grp-.....$")
	databaseTypeRegexp     = regexp.MustCompile("^dbt-.....$")
	databaseServerRegexp   = regexp.MustCompile("^dbs-.....$")
	databaseSnapshotRegexp = regexp.MustCompile("^dbi-.....$")
	loadBalancerRegexp     = regexp.MustCompile("^lba-.....$")
	zoneRegexp             = regexp.MustCompile("^(zon-.....$|gb1s?-[ab])$")
	serverTypeRegexp       = regexp.MustCompile("^typ-.....$")
	firewallPolicyRegexp   = regexp.MustCompile("^fwp-.....$")
	firewallRuleRegexp     = regexp.MustCompile("^fwr-.....$")
	interfaceRegexp        = regexp.MustCompile("^int-.....$")
	imageRegexp            = regexp.MustCompile("^img-.....$")
	volumeRegexp           = regexp.MustCompile("^vol-.....$")
	dnsNameRegexp          = regexp.MustCompile("^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$")
	unreadable             = map[string]bool{
		"deleted": true,
		"failed":  true,
	}
	validDatabaseEngines = []string{"mysql", "postgresql"}
)

func timeFromFloat(timeFloat float64) time.Time {
	sec, dec := math.Modf(timeFloat)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}

func hashString(
	v interface{},
) string {
	if str, ok := v.(string); ok {
		return userDataHashSum(str)
	}
	return ""
}

func userDataHashSum(userData string) string {
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it
	hash := sha1.Sum([]byte(userData))
	return hex.EncodeToString(hash[:])
}

func assignMap(d *schema.ResourceData, target **map[string]interface{}, index string) {
	if d.HasChange(index) {
		if attr, ok := d.GetOk(index); ok {
			temp := attr.(map[string]interface{})
			*target = &temp
		} else {
			temp := make(map[string]interface{})
			*target = &temp
		}
	}
}

func assignString(d *schema.ResourceData, target **string, index string) {
	if d.HasChange(index) {
		if *target == nil {
			var temp string
			*target = &temp
		}
		if attr, ok := d.GetOk(index); ok {
			**target = attr.(string)
		} else {
			**target = ""
		}
	}
}

func assignStringSet(d *schema.ResourceData, target *[]string, index string) {
	if d.HasChange(index) {
		*target = sliceFromStringSet(d, index)
	}
}

func sliceFromStringSet(d *schema.ResourceData, index string) []string {
	return expandStringValueList(d.Get(index).(*schema.Set).List())
}

// ExpandStringValueList takes the result of flatmap.Expand for an array of strings
// and returns a []string
func expandStringValueList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

func assignInt(d *schema.ResourceData, target **uint, index string) {
	if d.HasChange(index) {
		var temp uint
		if attr, ok := d.GetOk(index); ok {
			temp = uint(attr.(int))
		}
		*target = &temp
	}
}

func assignByte(d *schema.ResourceData, target **uint8, index string) {
	if d.HasChange(index) {
		var temp uint8
		if attr, ok := d.GetOk(index); ok {
			temp = uint8(attr.(int))
		}
		*target = &temp
	}
}

func assignBool(d *schema.ResourceData, target **bool, index string) {
	if d.HasChange(index) {
		var temp bool
		if attr, ok := d.GetOk(index); ok {
			temp = attr.(bool)
		}
		*target = &temp
	}
}

func assignEnum(d *schema.ResourceData, target encoding.TextUnmarshaler, index string) {
	if d.HasChange(index) {
		if attr, ok := d.GetOk(index); ok {
			target.UnmarshalText([]byte(attr.(string)))
		}
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

func compactZero[S ~[]E, E comparable](s S) S {
	var zero E
	i := 0
	for _, v := range s {
		if v != zero {
			s[i] = v
			i++
		}
	}
	return s[:i]
}

// ValidateCronString checks if the string is a valid cron layout
// An empty string is acceptable.
func ValidateCronString(v interface{}, name string) (warns []string, errors []error) {
	cronstr := v.(string)
	if cronstr == "" {
		return
	}
	if _, err := cronexpr.Parse(cronstr); err != nil {
		errors = append(errors, fmt.Errorf("%q not a valid Cron: %s", name, err))
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
//
//	header-field   = field-name ":" OWS field-value OWS
//	field-name     = token
//	tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
//	        "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
//	token = 1*tchar
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
// value instead.
func getenvWithDefault(key string, defaultValue string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultValue
}

// strSliceContains checks if a given string is contained in a slice
// When anybody asks why Go needs generics, here you go.
func strSliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// Check a JSON object is correct
func validateJSONObject(v interface{}, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	var j map[string]interface{}
	s := v.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a JSON object: %s", k, err)}
	}

	return nil, nil
}

func diffSuppressJSONObject(k, old, new string, d *schema.ResourceData) bool {
	if strSliceContains([]string{"{}", ""}, old) &&
		strSliceContains([]string{"{}", ""}, new) {
		return true
	}
	return false
}

// HashcodeString hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func HashcodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// StringIsValidFirewallTarget checks whether a string would
// pass the Iptables validation as a valid source or destination.
func stringIsValidFirewallTarget() schema.SchemaValidateFunc {
	return validation.Any(
		validation.StringInSlice([]string{"any"}, false),
		validation.StringMatch(serverRegexp, "must be a valid server ID"),
		validation.StringMatch(serverGroupRegexp, "must be a valid server group ID"),
		validation.StringMatch(loadBalancerRegexp, "must be a valid load balancer ID"),
		validation.IsCIDR,
		validation.IsIPAddress,
	)
}

// Get a list of server IDs from a list of servers
func serverIDListFromNodes(
	nodes []brightbox.Server,
) []string {
	nodeIds := make([]string, len(nodes))
	for i, node := range nodes {
		nodeIds[i] = node.ID
	}
	return nodeIds
}

// Create a ServerGroupMemberList from a list of servers
func serverGroupMemberListFromNodes(
	nodes []brightbox.Server,
) brightbox.ServerGroupMemberList {
	groupList := make([]brightbox.ServerGroupMember, len(nodes))
	for i, node := range nodes {
		groupList[i] = brightbox.ServerGroupMember{Server: node.ID}
	}
	return brightbox.ServerGroupMemberList{Servers: groupList}
}

// Returns the most recent item out of a slice of items with Dates
func mostRecent[S ~[]I, I brightbox.CreateDated](items S) *I {
	sortedItems := items
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAtUnix() > items[j].CreatedAtUnix()
	})
	return &sortedItems[0]
}

// filter returns a new slice with all elements from the from the
// input elements for which the provided predicate function returns true.
func filter[S ~[]T, T any](input S, pred func(T) bool) (output S) {
	for _, v := range input {
		if pred(v) {
			output = append(output, v)
		}
	}
	return output
}

// idList returns a list of identifiers from a list of identifiable objects
func idList[S ~[]O, O any](list S, identify func(v O) string) []string {
	ids := make([]string, len(list))
	for i, v := range list {
		ids[i] = identify(v)
	}
	return ids
}

// Generic Set Operations
// Set Difference: A - B
func Difference[S ~[]O, O comparable](a, b S) (diff S) {
	m := make(map[O]bool, len(b))

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func Intersection[S ~[]O, O comparable](a, b S) (intersect S) {
	m := make(map[O]bool, len(a))

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			intersect = append(intersect, item)
		}
	}
	return
}

func Union[S ~[]O, O comparable](a, b S) S {
	m := make(map[O]bool, len(a))

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}
