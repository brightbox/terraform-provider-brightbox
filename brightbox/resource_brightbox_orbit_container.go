package brightbox

import (
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	defaultContainerPermission = "storage"
)

func resourceBrightboxContainer() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Orbit Container resource",
		CreateContext: resourceBrightboxContainerCreate,
		ReadContext:   resourceBrightboxContainerRead,
		UpdateContext: resourceBrightboxContainerUpdate,
		DeleteContext: resourceBrightboxContainerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"bytes_used": {
				Description: "Number of bytes used by the container",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"container_read": {
				Description: "Who can read the container",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"container_sync_key": {
				Description: "Container sync key",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"container_sync_to": {
				Description: "Container to sync to",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"container_write": {
				Description: "Who can write to the container",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"created_at": {
				Description: "The time the container was created",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"history_location": {
				Description:   "History location",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"versions_location"},
			},

			"metadata": {
				Description: "Set of key/value metadata associated with the container",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateFunc: http1Keys,
			},

			"name": {
				Description:  "Name of the Container",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"object_count": {
				Description: "Number of objects in the container",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"storage_policy": {
				Description: "Any storage policy in place",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"versions_location": {
				Description:   "Versions Location",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"history_location"},
			},
		},
	}
}

func resourceBrightboxContainerCreate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).OrbitClient
	client.ProviderClient.Context = ctx

	log.Printf("[INFO] Creating Container")
	createOpts := getCreateContainerOptions(d)
	log.Printf("[DEBUG] Container create configuration: %#v", createOpts)
	currentContainerPath := containerPath(d)
	log.Printf("[DEBUG] Create path is: %s", currentContainerPath)
	container, err := containers.Create(client, currentContainerPath, createOpts).Extract()
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Container created with TransID %s", container.TransID)
	d.SetId(currentContainerPath)
	return resourceBrightboxContainerRead(ctx, d, meta)
}

func resourceBrightboxContainerDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).OrbitClient
	client.ProviderClient.Context = ctx

	log.Printf("[INFO] Deleting Container")
	container, err := containers.Delete(client, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Container deleted with TransID %s", container.TransID)
	return nil
}

func resourceBrightboxContainerUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).OrbitClient
	client.ProviderClient.Context = ctx

	log.Printf("[INFO] Updating Container")
	updateOpts := getUpdateContainerOptions(d)
	log.Printf("[INFO] Container update configuration: %#v", updateOpts)
	container, err := containers.Update(client, d.Id(), updateOpts).Extract()
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Container updated with TransID %s", container.TransID)
	return resourceBrightboxContainerRead(ctx, d, meta)
}

func resourceBrightboxContainerRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).OrbitClient
	client.ProviderClient.Context = ctx

	log.Printf("[DEBUG] Reading container: %s", d.Id())
	result := containers.Get(client, d.Id(), nil)
	getresult, err := result.Extract()
	if err != nil {
		log.Printf("[DEBUG] Checking if container is deleted")
		return diag.FromErr(CheckDeleted(d, result.Err, "container"))
	}
	log.Printf("[INFO] Container read with TransID %s", getresult.TransID)
	metadata, _ := result.ExtractMetadata()
	return setContainerAttributes(d, getresult, metadata)
}

func containerPath(
	d *schema.ResourceData,
) string {
	return d.Get("name").(string)
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
		dest[strings.ToLower(k)] = escapedString(v)
	}
	return dest
}

func setUnescapedString(d *schema.ResourceData, elem string, inputString string) error {
	temp, err := url.PathUnescape(inputString)
	if err != nil {
		return err
	}
	//lintignore:R001
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
	//lintignore:R001
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
	//lintignore:R001
	return d.Set(elem, dest)
}

func setContainerAttributes(
	d *schema.ResourceData,
	attr *containers.GetHeader,
	metadata map[string]string,
) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] Setting Container details from %#v", attr)
	if err := d.Set("name", d.Id()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("container_read", compactZero(attr.Read)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("container_write", compactZero(attr.Write)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("versions_location", attr.VersionsLocation); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("history_location", attr.HistoryLocation); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := setUnescapedStringMap(d, "metadata", metadata); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	//Computed
	if err := d.Set("storage_policy", attr.StoragePolicy); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("object_count", attr.ObjectCount); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("created_at", timeFromFloat(attr.Timestamp).Format(time.RFC3339)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("bytes_used", attr.BytesUsed); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	return diags
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

func getUpdateContainerOptions(
	d *schema.ResourceData,
) *containers.UpdateOpts {
	opts := &containers.UpdateOpts{}
	if d.HasChange("container_read") {
		temp := strings.Join(sliceFromStringSet(d, "container_read"), ",")
		opts.ContainerRead = &temp
	}
	if d.HasChange("container_write") {
		temp := strings.Join(sliceFromStringSet(d, "container_write"), ",")
		opts.ContainerWrite = &temp
	}
	if attr, ok := d.GetOk("metadata"); ok {
		opts.Metadata = escapedStringMetadata(attr)
	}
	if d.HasChange("metadata") {
		old, new := d.GetChange("metadata")
		opts.RemoveMetadata = removedMetadataKeys(old, new)
	}
	assignString(d, &opts.ContainerSyncTo, "container_sync_to")
	assignString(d, &opts.ContainerSyncKey, "container_sync_key")
	if attr, ok := d.GetOk("versions_location"); ok {
		if attr == "" {
			opts.RemoveVersionsLocation = "yup"
		} else {
			opts.VersionsLocation = attr.(string)
		}
	}
	if attr, ok := d.GetOk("history_location"); ok {
		if attr == "" {
			opts.RemoveHistoryLocation = "yup"
		} else {
			opts.HistoryLocation = attr.(string)
		}
	}
	return opts
}

func getCreateContainerOptions(
	d *schema.ResourceData,
) *containers.CreateOpts {
	opts := &containers.CreateOpts{}
	opts.ContainerRead = strings.Join(sliceFromStringSet(d, "container_read"), ",")
	opts.ContainerWrite = strings.Join(sliceFromStringSet(d, "container_write"), ",")
	if attr, ok := d.GetOk("metadata"); ok {
		opts.Metadata = escapedStringMetadata(attr)
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
