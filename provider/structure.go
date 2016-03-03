package brightbox

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/hashicorp/terraform/helper/schema"
)

func hash_string(
	v interface{},
) string {
	switch v.(type) {
	case string:
		hash := sha1.Sum([]byte(v.(string)))
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

func assign_string(d *schema.ResourceData, target **string, index string) {
	if d.HasChange(index) {
		var temp string
		if attr, ok := d.GetOk(index); ok {
			temp = attr.(string)
		}
		*target = &temp
	}
}

func assign_string_set(d *schema.ResourceData, target **[]string, index string) {
	if d.HasChange(index) {
		var temp []string
		if attr := d.Get(index).(*schema.Set); attr.Len() > 0 {
			for _, v := range attr.List() {
				temp = append(temp, v.(string))
			}
		}
		*target = &temp
	}
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
