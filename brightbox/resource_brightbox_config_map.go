package brightbox

import (
	"fmt"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceBrightboxConfigMap() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxConfigMapCreate,
		Read:   resourceBrightboxConfigMapRead,
		Update: resourceBrightboxConfigMapUpdate,
		Delete: resourceBrightboxConfigMapDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "User editable label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"data": {
				Description: "keys/values making up the ConfigMap",
				Required:    true,
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceBrightboxConfigMapCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Config Map")
	configMapOpts := &brightbox.ConfigMapOptions{}
	err := addUpdateableConfigMapOptions(d, configMapOpts)
	if err != nil {
		return err
	}

	configMap, err := client.CreateConfigMap(configMapOpts)
	if err != nil {
		return fmt.Errorf("Error creating Config Map: %s", err)
	}

	d.SetId(configMap.Id)

	return setConfigMapAttributes(d, configMap)
}

func resourceBrightboxConfigMapRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	configMap, err := client.ConfigMap(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Config Map not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Config Map details: %s", err)
	}

	return setConfigMapAttributes(d, configMap)
}

func resourceBrightboxConfigMapDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Deleting Config Map %s", d.Id())
	err := client.DestroyConfigMap(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Config Map (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxConfigMapUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	configMapOpts := &brightbox.ConfigMapOptions{
		Id: d.Id(),
	}
	err := addUpdateableConfigMapOptions(d, configMapOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Config Map update configuration: %#v", configMapOpts)

	configMap, err := client.UpdateConfigMap(configMapOpts)
	if err != nil {
		return fmt.Errorf("Error updating Config Map (%s): %s", configMapOpts.Id, err)
	}

	return setConfigMapAttributes(d, configMap)
}

func addUpdateableConfigMapOptions(
	d *schema.ResourceData,
	opts *brightbox.ConfigMapOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_map(d, &opts.Data, "data")
	return nil
}

func setConfigMapAttributes(
	d *schema.ResourceData,
	configMap *brightbox.ConfigMap,
) error {
	d.Set("name", configMap.Name)
	d.Set("data", configMap.Data)
	return nil
}
