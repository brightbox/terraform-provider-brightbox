package brightbox

import (
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	default_container_permission = "storage"
)

func resourceBrightboxContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxContainerCreate,
		Read:   resourceBrightboxContainerRead,
		Update: resourceBrightboxContainerUpdate,
		Delete: resourceBrightboxContainerDelete,
		//Exists: resourceBrightboxContainerExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateFunc: http1Keys,
			},
			"container_read": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"container_write": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"container_sync_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"container_sync_to": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"versions_location": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"history_location"},
			},
			"history_location": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"versions_location"},
			},
			"object_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"bytes_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"storage_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBrightboxContainerCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).OrbitClient

	log.Printf("[INFO] Creating Container")
	createOpts := getCreateContainerOptions(d)
	log.Printf("[DEBUG] Container create configuration: %#v", createOpts)
	container_path := containerPath(d)
	log.Printf("[DEBUG] Create path is: %s", container_path)
	container, err := containers.Create(client, container_path, createOpts).Extract()
	if err != nil {
		return err
	}
	log.Printf("[INFO] Container created with TransID %s", container.TransID)
	d.SetId(container_path)
	return resourceBrightboxContainerRead(d, meta)
}

func resourceBrightboxContainerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).OrbitClient

	log.Printf("[INFO] Deleting Container")
	container, err := containers.Delete(client, d.Id()).Extract()
	if err != nil {
		return err
	}
	log.Printf("[INFO] Container deleted with TransID %s", container.TransID)
	return nil
}

func resourceBrightboxContainerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).OrbitClient

	log.Printf("[INFO] Updating Container")
	updateOpts := getUpdateContainerOptions(d)
	log.Printf("[INFO] Container update configuration: %#v", updateOpts)
	container, err := containers.Update(client, d.Id(), updateOpts).Extract()
	if err != nil {
		return err
	}
	log.Printf("[INFO] Container updated with TransID %s", container.TransID)
	return resourceBrightboxContainerRead(d, meta)
}

func resourceBrightboxContainerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).OrbitClient

	log.Printf("[DEBUG] Reading container: %s", d.Id())
	result := containers.Get(client, d.Id(), nil)
	getresult, err := result.Extract()
	if err != nil {
		log.Printf("[DEBUG] Checking if container is deleted")
		return CheckDeleted(d, result.Err, "container")
	}
	log.Printf("[INFO] Container read with TransID %s", getresult.TransID)
	metadata, _ := result.ExtractMetadata()
	return setContainerAttributes(d, getresult, metadata)
}

//func resourceBrightboxContainerExists(
//	d *schema.ResourceData,
//	meta interface{},
//) (bool, error) {
//	client := meta.(*CompositeClient).OrbitClient
//
//	log.Printf("[DEBUG] Checking if container exists: %s", d.Id())
//	getresult := containers.Get(client, d.Id(), nil)
//	return getresult.Err == nil, getresult.Err
//}

func containerPath(
	d *schema.ResourceData,
) string {
	return d.Get("name").(string)
}

func setUnescapedString(d *schema.ResourceData, elem string, inputString string) error {
	temp, err := url.PathUnescape(inputString)
	if err != nil {
		return err
	}
	return d.Set(elem, temp)
}

func setUnescapedStringSet(d *schema.ResourceData, elem string, inputStringSet []string) error {
	var tempSet []string
	for _, str := range inputStringSet {
		if str != "" {
			temp, err := url.PathUnescape(str)
			if err != nil {
				return err
			}
			tempSet = append(tempSet, temp)
		}
	}
	return d.Set(elem, tempSet)
}

func setUnescapedStringMap(d *schema.ResourceData, elem string, inputMap map[string]string) error {
	dest := make(map[string]string)
	source := inputMap
	for k, v := range source {
		temp, err := url.PathUnescape(v)
		if err != nil {
			return err
		}
		dest[strings.ToLower(k)] = temp
	}
	return d.Set(elem, dest)
}

func setContainerAttributes(
	d *schema.ResourceData,
	attr *containers.GetHeader,
	metadata map[string]string,
) error {
	log.Printf("[DEBUG] Setting Container details from %#v", attr)
	if err := setUnescapedString(d, "name", d.Id()); err != nil {
		return err
	}
	if err := setUnescapedStringSet(d, "container_read", attr.Read); err != nil {
		return err
	}
	if err := setUnescapedStringSet(d, "container_write", attr.Write); err != nil {
		return err
	}
	if err := setUnescapedString(d, "versions_location", attr.VersionsLocation); err != nil {
		return err
	}
	if err := setUnescapedString(d, "history_location", attr.HistoryLocation); err != nil {
		return err
	}
	if err := setUnescapedStringMap(d, "metadata", metadata); err != nil {
		return err
	}
	//Computed
	if err := setUnescapedString(d, "storage_policy", attr.StoragePolicy); err != nil {
		return err
	}
	if err := d.Set("object_count", attr.ObjectCount); err != nil {
		return err
	}
	if err := d.Set("created_at", attr.Date.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("bytes_used", attr.BytesUsed); err != nil {
		return err
	}
	return nil
}

func getUpdateContainerOptions(
	d *schema.ResourceData,
) *containers.UpdateOpts {
	opts := &containers.UpdateOpts{}
	opts.ContainerRead = strings.Join(escapedStringList(map_from_string_set(d, "container_read")), ",")
	opts.ContainerWrite = strings.Join(escapedStringList(map_from_string_set(d, "container_write")), ",")
	if attr, ok := d.GetOk("metadata"); ok {
		opts.Metadata = escapedStringMetadata(attr)
	}
	if attr, ok := d.GetOk("container_sync_to"); ok {
		opts.ContainerSyncTo = escapedString(attr)
	}
	if attr, ok := d.GetOk("container_sync_key"); ok {
		opts.ContainerSyncKey = escapedString(attr)
	}
	if attr, ok := d.GetOk("versions_location"); ok {
		if attr == "" {
			opts.RemoveVersionsLocation = "yup"
		} else {
			opts.VersionsLocation = escapedString(attr)
		}
	}
	if attr, ok := d.GetOk("history_location"); ok {
		if attr == "" {
			opts.RemoveHistoryLocation = "yup"
		} else {
			opts.HistoryLocation = escapedString(attr)
		}
	}
	return opts
}

func getCreateContainerOptions(
	d *schema.ResourceData,
) *containers.CreateOpts {
	opts := &containers.CreateOpts{}
	opts.ContainerRead = strings.Join(escapedStringList(map_from_string_set(d, "container_read")), ",")
	opts.ContainerWrite = strings.Join(escapedStringList(map_from_string_set(d, "container_write")), ",")
	if attr, ok := d.GetOk("metadata"); ok {
		opts.Metadata = escapedStringMetadata(attr)
	}
	if attr, ok := d.GetOk("container_sync_to"); ok {
		opts.ContainerSyncTo = escapedString(attr)
	}
	if attr, ok := d.GetOk("container_sync_key"); ok {
		opts.ContainerSyncKey = escapedString(attr)
	}
	if attr, ok := d.GetOk("versions_location"); ok {
		opts.VersionsLocation = escapedString(attr)
	}
	if attr, ok := d.GetOk("history_location"); ok {
		opts.HistoryLocation = escapedString(attr)
	}
	return opts
}
