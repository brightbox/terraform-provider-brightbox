package brightbox

import (
	"fmt"
	"log"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var blankDatabaseServerOpts = brightbox.DatabaseServerOptions{}

func resourceBrightboxDatabaseServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxDatabaseServerCreate,
		Read:   resourceBrightboxDatabaseServerRead,
		Update: resourceBrightboxDatabaseServerUpdate,
		Delete: resourceBrightboxDatabaseServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"maintenance_weekday": {
				Description: "Numerical index of weekday (0 is Sunday, 1 is Monday...) to set when automatic updates may be performed",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"maintenance_hour": {
				Description: "Number representing 24hr time start of maintenance window hour for x:00-x:59 (0-23)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"database_engine": {
				Description: "The DBMS engine of the Database Server",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"database_version": {
				Description: "The version of the given engine in use",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"database_type": {
				Description: "ID of the database type to use",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"allow_access": {
				Description: "An array of resources allowed to access the database. Accepted values include `any`, `IPv4 address`, `server identifier`, `server group identifier`",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				MinItems:    1,
				Set:         schema.HashString,
			},
			"snapshot": {
				Description: "Identifier for an SQL snapshot to use as the basis of the new instance. Creates and restores the database from the snapshot",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"snapshots_schedule": {
				Description:  "Crontab pattern for scheduled snapshots. Must be at least hourly",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: ValidateCronString,
			},
			"snapshots_schedule_next_at": {
				Description: "time in UTC when next approximate scheduled snapshot will be run",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"zone": {
				Description: "ID of the zone the database server is in",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"admin_username": {
				Description: "Initial username required to login",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"admin_password": {
				Description: "Initial password required to login, only available at creation or following a password reset request",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"status": {
				Description: "State the database server is in",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"locked": {
				Description: "Initial password required to login, only available at creation or following a password reset request",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func setDatabaseServerAttributes(
	d *schema.ResourceData,
	databaseServer *brightbox.DatabaseServer,
) {
	d.Set("name", databaseServer.Name)
	d.Set("description", databaseServer.Description)
	d.Set("status", databaseServer.Status)
	d.Set("locked", databaseServer.Locked)
	d.Set("database_engine", databaseServer.DatabaseEngine)
	d.Set("database_version", databaseServer.DatabaseVersion)
	d.Set("database_type", databaseServer.DatabaseServerType.Id)
	d.Set("admin_username", databaseServer.AdminUsername)
	d.Set("maintenance_weekday", databaseServer.MaintenanceWeekday)
	d.Set("maintenance_hour", databaseServer.MaintenanceHour)
	d.Set("snapshots_schedule", databaseServer.SnapshotsSchedule)
	d.Set("snapshots_schedule_next_at", databaseServer.SnapshotsScheduleNextAt.Format(time.RFC3339))
	d.Set("zone", databaseServer.Zone.Handle)
}

func setAllowAccessAttribute(
	d *schema.ResourceData,
	databaseServer *brightbox.DatabaseServer,
) error {
	if err := d.Set("allow_access", schema.NewSet(schema.HashString, flatten_string_slice(databaseServer.AllowAccess))); err != nil {
		return fmt.Errorf("error setting allow_access: %s", err)
	}
	return nil
}

func databaseServerStateRefresh(client *brightbox.Client, databaseServerID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		databaseServer, err := client.DatabaseServer(databaseServerID)
		if err != nil {
			log.Printf("Error on Database Server State Refresh: %s", err)
			return nil, "", err
		}
		return databaseServer, databaseServer.Status, nil
	}
}

func resourceBrightboxDatabaseServerCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient
	err := createDatabaseServer(d, client)
	if err != nil {
		return err
	}
	databaseServerOpts := getBlankDatabaseServerOpts()
	databaseServerOpts.AllowAccess = map_from_string_set(d, "allow_access")
	return updateDatabaseServerAttributes(d, meta, client, databaseServerOpts)
}

func createDatabaseServer(d *schema.ResourceData, client *brightbox.Client) error {
	log.Printf("[DEBUG] Database Server create called")
	databaseServerOpts := getBlankDatabaseServerOpts()
	err := addUpdateableDatabaseServerOptions(d, databaseServerOpts)
	if err != nil {
		return err
	}
	engine := &databaseServerOpts.Engine
	assign_string(d, &engine, "database_engine")
	version := &databaseServerOpts.Version
	assign_string(d, &version, "database_version")
	databaseType := &databaseServerOpts.DatabaseType
	assign_string(d, &databaseType, "database_type")
	snapshot := &databaseServerOpts.Snapshot
	assign_string(d, &snapshot, "snapshot")
	zone := &databaseServerOpts.Zone
	assign_string(d, &zone, "zone")
	log.Printf("[DEBUG] Database Server create configuration %#v", databaseServerOpts)
	outputDatabaseServerOptions(databaseServerOpts)
	databaseServer, err := client.CreateDatabaseServer(databaseServerOpts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(databaseServer.Id)

	if databaseServer.AdminPassword == "" {
		log.Printf("[WARN] No password returned for Cloud SQL server %s", databaseServer.Id)
	} else {
		d.Set("admin_password", databaseServer.AdminPassword)
	}

	locked := d.Get("locked").(bool)
	log.Printf("[INFO] Setting lock state to %v", locked)
	if err := setLockState(client, locked, databaseServer); err != nil {
		return err
	}

	log.Printf("[INFO] Waiting for Database Server (%s) to become available", d.Id())
	stateConf := resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    databaseServerStateRefresh(client, databaseServer.Id),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	activeDatabaseServer, err := stateConf.WaitForState()
	if err != nil {
		return err
	}
	setDatabaseServerAttributes(d, activeDatabaseServer.(*brightbox.DatabaseServer))
	return nil
}

func resourceBrightboxDatabaseServerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	// Create/Update Database
	databaseServerOpts := getBlankDatabaseServerOpts()
	err := addUpdateableDatabaseServerOptions(d, databaseServerOpts)
	if err != nil {
		return err
	}
	assign_string_set(d, &databaseServerOpts.AllowAccess, "allow_access")
	log.Printf("[DEBUG] Database Server update configuration %#v", databaseServerOpts)
	outputDatabaseServerOptions(databaseServerOpts)
	return updateDatabaseServerAttributes(d, meta, client, databaseServerOpts)
}

func updateDatabaseServer(
	client *brightbox.Client,
	databaseServerOpts *brightbox.DatabaseServerOptions,
) (*brightbox.DatabaseServer, error) {
	log.Printf("[DEBUG] Database Server update configuration %#v", databaseServerOpts)
	outputDatabaseServerOptions(databaseServerOpts)
	databaseServer, err := client.UpdateDatabaseServer(databaseServerOpts)
	if err != nil {
		return nil, fmt.Errorf("Error updating databaseServer: %s", err)
	}
	return databaseServer, nil
}

func getBlankDatabaseServerOpts() *brightbox.DatabaseServerOptions {
	temp := blankDatabaseServerOpts
	return &temp
}

func updateDatabaseServerAttributes(
	d *schema.ResourceData,
	meta interface{},
	client *brightbox.Client,
	databaseServerOpts *brightbox.DatabaseServerOptions,
) error {
	if cmp.Equal(*databaseServerOpts, blankDatabaseServerOpts) && !d.HasChange("locked") {
		return fmt.Errorf("[ERROR] No database update changes detected for %s", d.Id())
	}
	databaseServerOpts.Id = d.Id()
	databaseServer, err := updateDatabaseServer(client, databaseServerOpts)
	if err != nil {
		return err
	}

	if d.HasChange("locked") {
		locked := d.Get("locked").(bool)
		log.Printf("[INFO] Setting lock state to %v", locked)
		if err := setLockState(client, locked, databaseServer); err != nil {
			return err
		}
		return resourceBrightboxDatabaseServerRead(d, meta)
	}

	setDatabaseServerAttributes(d, databaseServer)
	return setAllowAccessAttribute(d, databaseServer)
}

func resourceBrightboxDatabaseServerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient
	log.Printf("[DEBUG] Database Server read called for %s", d.Id())
	databaseServer, err := client.DatabaseServer(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Database Server details: %s", err)
	}
	if databaseServer.Status == "deleted" {
		log.Printf("[WARN] Database Server not found, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}
	setDatabaseServerAttributes(d, databaseServer)
	return setAllowAccessAttribute(d, databaseServer)
}

func resourceBrightboxDatabaseServerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Database Server delete called for %s", d.Id())
	err := client.DestroyDatabaseServer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Database Server: %s", err)
	}
	stateConf := resource.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    databaseServerStateRefresh(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	return nil
}

func addUpdateableDatabaseServerOptions(
	d *schema.ResourceData,
	opts *brightbox.DatabaseServerOptions,
) error {
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Description, "description")
	assign_int(d, &opts.MaintenanceWeekday, "maintenance_weekday")
	assign_int(d, &opts.MaintenanceHour, "maintenance_hour")
	assign_string(d, &opts.SnapshotsSchedule, "snapshots_schedule")
	return nil
}

func outputDatabaseServerOptions(opts *brightbox.DatabaseServerOptions) {
	if opts.Name != nil {
		log.Printf("[DEBUG] Database Server Name %v", *opts.Name)
	}
	if opts.Description != nil {
		log.Printf("[DEBUG] Database Server Description %v", *opts.Description)
	}
	if opts.Engine != "" {
		log.Printf("[DEBUG] Database Server Engine %v", opts.Engine)
	}
	if opts.Version != "" {
		log.Printf("[DEBUG] Database Server Version %v", opts.Version)
	}
	if opts.DatabaseType != "" {
		log.Printf("[DEBUG] Database Server Type %v", opts.DatabaseType)
	}
	if opts.AllowAccess != nil {
		log.Printf("[DEBUG] Database Server AllowAccess %#v", opts.AllowAccess)
	}
	if opts.Snapshot != "" {
		log.Printf("[DEBUG] Database Server Snapshot %v", opts.Snapshot)
	}
	if opts.SnapshotsSchedule != nil {
		log.Printf("[DEBUG] Database Server Snapshots Schedule %v", *opts.SnapshotsSchedule)
	}
	if opts.Zone != "" {
		log.Printf("[DEBUG] Database Server Zone %v", opts.Zone)
	}
	if opts.MaintenanceWeekday != nil {
		log.Printf("[DEBUG] Database Server MaintenanceWeekday %v", *opts.MaintenanceWeekday)
	}
	if opts.MaintenanceHour != nil {
		log.Printf("[DEBUG] Database Server MaintenanceHour %v", *opts.MaintenanceHour)
	}
}
