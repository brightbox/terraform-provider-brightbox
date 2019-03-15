package brightbox

import (
	"log"
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
		Exists: resourceBrightboxContainerExists,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"container_read": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"container_write": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"container_sync_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"container_sync_to": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_type": {
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
			"content_length": {
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
		if strings.HasPrefix(err.Error(), "missing resource:") {
			log.Printf("[WARN] Container not found, removing from state")
			d.SetId("")
			return nil
		}
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
		return err
	}
	log.Printf("[INFO] Container read with TransID %s", getresult.TransID)
	metadata, _ := result.ExtractMetadata()
	return setContainerAttributes(d, getresult, metadata)
}

func resourceBrightboxContainerExists(
	d *schema.ResourceData,
	meta interface{},
) (bool, error) {
	client := meta.(*CompositeClient).OrbitClient

	getresult := containers.Get(client, d.Id(), nil)
	return getresult.Err == nil, getresult.Err
}

func containerPath(
	d *schema.ResourceData,
) string {
	return d.Get("name").(string)
}

func setContainerAttributes(
	d *schema.ResourceData,
	attr *containers.GetHeader,
	metadata map[string]string,
) error {
	log.Printf("[DEBUG] Setting Container details")
	if err := d.Set("name", d.Id()); err != nil {
		return err
	}
	if err := d.Set("content_type", attr.ContentType); err != nil {
		return err
	}
	if err := d.Set("container_read", attr.Read); err != nil {
		return err
	}
	if err := d.Set("container_write", attr.Write); err != nil {
		return err
	}
	if err := d.Set("versions_location", attr.VersionsLocation); err != nil {
		return err
	}
	if err := d.Set("history_location", attr.HistoryLocation); err != nil {
		return err
	}
	if err := d.Set("metadata", metadata); err != nil {
		return err
	}
	//Computed
	if err := d.Set("storage_policy", attr.StoragePolicy); err != nil {
		return err
	}
	if err := d.Set("object_count", attr.ObjectCount); err != nil {
		return err
	}
	if err := d.Set("created_at", attr.Date.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("content_length", attr.ContentLength); err != nil {
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
	opts.ContainerRead = strings.Join(map_from_string_set(d, "container_read"), ",")
	opts.ContainerWrite = strings.Join(map_from_string_set(d, "container_write"), ",")
	if attr, ok := d.GetOk("metadata"); ok {
		opts.Metadata = attr.(map[string]string)
	}
	if attr, ok := d.GetOk("content_type"); ok {
		opts.ContentType = attr.(string)
	}
	if attr, ok := d.GetOk("container_sync_to"); ok {
		opts.ContainerSyncTo = attr.(string)
	}
	if attr, ok := d.GetOk("container_sync_key"); ok {
		opts.ContainerSyncKey = attr.(string)
	}
	if attr, ok := d.GetOk("versions_location"); ok {
		if attr == "" {
			opts.RemoveVersionsLocation = "yup"
		}
		opts.VersionsLocation = attr.(string)
	}
	if attr, ok := d.GetOk("history_location"); ok {
		if attr == "" {
			opts.RemoveHistoryLocation = "yup"
		}
		opts.HistoryLocation = attr.(string)
	}
	return opts
}

func getCreateContainerOptions(
	d *schema.ResourceData,
) *containers.CreateOpts {
	opts := &containers.CreateOpts{}
	opts.ContainerRead = strings.Join(map_from_string_set(d, "container_read"), ",")
	opts.ContainerWrite = strings.Join(map_from_string_set(d, "container_write"), ",")
	if attr, ok := d.GetOk("metadata"); ok {
		source := attr.(map[string]interface{})
		dest := make(map[string]string, len(source))
		for i := range source {
			dest[i] = source[i].(string)
		}
		opts.Metadata = dest
	}
	if attr, ok := d.GetOk("content_type"); ok {
		opts.ContentType = attr.(string)
	}
	if attr, ok := d.GetOk("container_sync_to"); ok {
		opts.ContainerSyncTo = attr.(string)
	}
	if attr, ok := d.GetOk("container_sync_key"); ok {
		opts.ContainerSyncKey = attr.(string)
	}
	if attr, ok := d.GetOk("versions_location"); ok {
		opts.VersionsLocation = attr.(string)
	}
	if attr, ok := d.GetOk("history_location"); ok {
		opts.HistoryLocation = attr.(string)
	}
	return opts
}

func fromId(path string) (string, string) {
	elem := strings.SplitN(path, "/", 2)
	return elem[0], elem[1]
}
