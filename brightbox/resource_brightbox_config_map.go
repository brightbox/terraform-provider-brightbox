package brightbox

import (
	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func resourceBrightboxConfigMap() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Brightbox Config Map resource",
		CreateContext: resourceBrightboxCreate(
			(*brightbox.Client).CreateConfigMap,
			"Config Map",
			addUpdateableConfigMapOptions,
			setConfigMapAttributes,
		),
		ReadContext: resourceBrightboxRead(
			(*brightbox.Client).ConfigMap,
			"Config Map",
			setConfigMapAttributes,
		),
		UpdateContext: resourceBrightboxUpdate(
			(*brightbox.Client).UpdateConfigMap,
			"Config Map",
			configMapFromID,
			addUpdateableConfigMapOptions,
			setConfigMapAttributes,
		),
		DeleteContext: resourceBrightboxDelete(
			(*brightbox.Client).DestroyConfigMap,
			"Config Map",
		),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"data": {
				Description: "keys/values making up the ConfigMap",
				Required:    true,
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateFunc:     validateJSONObject,
					DiffSuppressFunc: diffSuppressJSONObject,
					StateFunc: func(v interface{}) string {
						json, _ := structure.NormalizeJsonString(v)
						return json
					},
				},
			},

			"name": {
				Description: "User editable label",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func configMapFromID(
	id string,
) *brightbox.ConfigMapOptions {
	return &brightbox.ConfigMapOptions{
		ID: id,
	}
}

func addUpdateableConfigMapOptions(
	d *schema.ResourceData,
	opts *brightbox.ConfigMapOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignMap(d, &opts.Data, "data")
	return nil
}

func setConfigMapAttributes(
	d *schema.ResourceData,
	configMap *brightbox.ConfigMap,
) diag.Diagnostics {
	var diags diag.Diagnostics
	d.SetId(configMap.ID)
	err := d.Set("name", configMap.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("data", configMap.Data)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	return diags
}
