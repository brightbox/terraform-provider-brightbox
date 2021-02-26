package brightbox

import (
	"context"
	"fmt"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func resourceBrightboxConfigMap() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Config Map resource",
		CreateContext: resourceBrightboxConfigMapCreate,
		ReadContext:   resourceBrightboxConfigMapRead,
		UpdateContext: resourceBrightboxConfigMapUpdate,
		DeleteContext: resourceBrightboxConfigMapDelete,
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

func resourceBrightboxConfigMapCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Config Map")
	configMapOpts := &brightbox.ConfigMapOptions{}
	err := addUpdateableConfigMapOptions(d, configMapOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Config Map create configuration %#v", configMapOpts)
	configMap, err := client.CreateConfigMap(configMapOpts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating Config Map: %s", err))
	}

	d.SetId(configMap.Id)

	return setConfigMapAttributes(d, configMap)
}

func resourceBrightboxConfigMapRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	configMap, err := client.ConfigMap(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Config Map not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("Error retrieving Config Map details: %s", err))
	}

	return setConfigMapAttributes(d, configMap)
}

func resourceBrightboxConfigMapDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Deleting Config Map %s", d.Id())
	err := client.DestroyConfigMap(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error deleting Config Map (%s): %s", d.Id(), err))
	}
	return nil
}

func resourceBrightboxConfigMapUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	configMapOpts := &brightbox.ConfigMapOptions{
		Id: d.Id(),
	}
	err := addUpdateableConfigMapOptions(d, configMapOpts)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Config Map update configuration: %#v", configMapOpts)

	configMap, err := client.UpdateConfigMap(configMapOpts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error updating Config Map (%s): %s", configMapOpts.Id, err))
	}

	return setConfigMapAttributes(d, configMap)
}

func addUpdateableConfigMapOptions(
	d *schema.ResourceData,
	opts *brightbox.ConfigMapOptions,
) error {
	assignString(d, &opts.Name, "name")
	assignMap(d, &opts.Data, "data")
	return nil
}

func setConfigMapAttributes(
	d *schema.ResourceData,
	configMap *brightbox.ConfigMap,
) diag.Diagnostics {
	var diags diag.Diagnostics
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
