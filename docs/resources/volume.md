# brightbox\_volume Resource

Provides a Brightbox Volume resource. This can be used to create,
modify, and delete Volumes.

## Example Usage

```hcl
resource "brightbox_volume" "boot_disk" {
  name = "Terraform web server example boot disk"
  image = data.brightbox_image.ubuntu_lts.id
  size = 61440
}

resource "brightbox_volume" "data_disk" {
  name = "Terraform web server example data disk"
  size = 40960
  serial = "TWS_DATA_DISK"
  filesystem_type = "xfs"
  filesystem_label = "data_area"
  server = brightbox_server.web.id
}

resource "brightbox_server" "web" {
  name  = "Terraform web server example"
  volume = brightbox_volume.boot_disk.id
  type  = data.brightbox_server_type.nbs_type.id
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A label assigned to the Volume
* `description` - (Optional) Verbose Description of this volume
* `encrypted` - (Optional) True if the volume is encrypted
* `filesystem_label` - (Optional) Label given to the filesystem on the volume. Up to 12 characters.
* `filesystem_type` - (Optional) Format of the filesystem on the volume. Either `ext4` or `xfs`. One of `image`, `filesystem_type` or `source` is required.
* `image` - (Optional) Image used to create the volume. One of `image`, `filesystem_type` or `source` is required.
* `serial` - (Optional) Volume Serial Number. Up to 20 characters.
* `server` - (Optional) The ID of the server this volume should be attached to.
* `size` - (Optional) Disk size in megabytes
* `source` - (Optional) The ID of the source volume for this image. Defaults to the blank disk.


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Volume
* `status` - The current state of the volume
* `source_type` - Source type for this image. One of `image`, `volume` or `raw`
* `storage_type` - Storage type for this volume. Either `local` or `network`


## Import

Volumes can be imported using the volume `id`, e.g.

```
terraform import brightbox_volume.default vol-ok8vw
```

<a id="timeouts"></a>
## Timeouts

`brightbox_volume` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `5 minutes`) Used for Creating Volumes
- `delete` - (Default `5 minutes`) Used for Deleting Volumes
