package brightbox

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brightbox/gobrightbox"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

var blank_database_server_opts = brightbox.DatabaseServerOptions{}

func resourceBrightboxDatabaseServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxDatabaseServerCreate,
		Read:   resourceBrightboxDatabaseServerRead,
		Update: resourceBrightboxDatabaseServerUpdate,
		Delete: resourceBrightboxDatabaseServerDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"maintenance_weekday": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"maintenance_hour": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"database_engine": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"database_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"database_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"allow_access": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				MinItems: 1,
				Set:      schema.HashString,
			},
			"snapshot": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"admin_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_password": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func setDatabaseServerAttributes(
	d *schema.ResourceData,
	database_server *brightbox.DatabaseServer,
) {
	d.Set("name", database_server.Name)
	d.SetPartial("name")
	d.Set("Description", database_server.Description)
	d.SetPartial("description")
	d.Set("status", database_server.Status)
	d.SetPartial("status")
	d.Set("locked", database_server.Locked)
	d.SetPartial("locked")
	d.Set("database_engine", database_server.DatabaseEngine)
	d.SetPartial("database_engine")
	d.Set("database_version", database_server.DatabaseVersion)
	d.SetPartial("database_version")
	d.Set("database_type", database_server.DatabaseServerType.Id)
	d.SetPartial("database_type")
	d.Set("admin_username", database_server.AdminUsername)
	d.SetPartial("admin_username")
	d.Set("maintenance_weekday", database_server.MaintenanceWeekday)
	d.SetPartial("maintenance_weekday")
	d.Set("maintenance_hour", database_server.MaintenanceHour)
	d.SetPartial("maintenance_hour")
	d.Set("zone", database_server.Zone.Handle)
	d.SetPartial("zone")
}

func setAllowAccessAttribute(
	d *schema.ResourceData,
	database_server *brightbox.DatabaseServer,
) {
	d.Set("allow_access", schema.NewSet(schema.HashString, flatten_string_slice(database_server.AllowAccess)))
	d.SetPartial("allow_access")
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
	client := meta.(*CompositeClient).ApiClient
	err := createDatabaseServer(d, client)
	if err != nil {
		return err
	}
	database_server_opts := getBlankDatabaseServerOpts()
	database_server_opts.AllowAccess = map_from_string_set(d, "allow_access")
	return updateDatabaseServerAttributes(d, client, database_server_opts)
}

func createDatabaseServer(d *schema.ResourceData, client *brightbox.Client) error {
	log.Printf("[DEBUG] Database Server create called")
	database_server_opts := getBlankDatabaseServerOpts()
	err := addUpdateableDatabaseServerOptions(d, database_server_opts)
	if err != nil {
		return err
	}
	engine := &database_server_opts.Engine
	assign_string(d, &engine, "database_engine")
	version := &database_server_opts.Version
	assign_string(d, &version, "database_version")
	databaseType := &database_server_opts.DatabaseType
	assign_string(d, &databaseType, "database_type")
	snapshot := &database_server_opts.Snapshot
	assign_string(d, &snapshot, "snapshot")
	zone := &database_server_opts.Zone
	assign_string(d, &zone, "zone")
	log.Printf("[DEBUG] Database Server create configuration %#v", database_server_opts)
	output_database_server_options(database_server_opts)
	database_server, err := client.CreateDatabaseServer(database_server_opts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}
	log.Printf("[DEBUG] Setting Partial")
	d.Partial(true)
	d.SetId(database_server.Id)
	if database_server.AdminPassword == "" {
		log.Printf("[WARN] No password returned for Cloud SQL server %s", database_server.Id)
	} else {
		d.Set("admin_password", database_server.AdminPassword)
	}
	log.Printf("[INFO] Waiting for Database Server (%s) to become available", d.Id())
	stateConf := resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    databaseServerStateRefresh(client, database_server.Id),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	active_database_server, err := stateConf.WaitForState()
	if err != nil {
		return err
	}
	d.SetPartial("admin_password")
	setDatabaseServerAttributes(d, active_database_server.(*brightbox.DatabaseServer))
	return nil
}

func resourceBrightboxDatabaseServerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient
	log.Printf("[DEBUG] Setting Partial")
	d.Partial(true)
	// Create/Update Database
	database_server_opts := getBlankDatabaseServerOpts()
	err := addUpdateableDatabaseServerOptions(d, database_server_opts)
	if err != nil {
		return err
	}
	assign_string_set(d, &database_server_opts.AllowAccess, "allow_access")
	log.Printf("[DEBUG] Database Server update configuration %#v", database_server_opts)
	output_database_server_options(database_server_opts)
	return updateDatabaseServerAttributes(d, client, database_server_opts)
}

func updateDatabaseServer(
	client *brightbox.Client,
	database_server_opts *brightbox.DatabaseServerOptions,
) (*brightbox.DatabaseServer, error) {
	log.Printf("[DEBUG] Database Server update configuration %#v", database_server_opts)
	output_database_server_options(database_server_opts)
	database_server, err := client.UpdateDatabaseServer(database_server_opts)
	if err != nil {
		return nil, fmt.Errorf("Error updating database_server: %s", err)
	}
	return database_server, nil
}

func getBlankDatabaseServerOpts() *brightbox.DatabaseServerOptions {
	temp := blank_database_server_opts
	return &temp
}

func updateDatabaseServerAttributes(
	d *schema.ResourceData,
	client *brightbox.Client,
	database_server_opts *brightbox.DatabaseServerOptions,
) error {
	if cmp.Equal(*database_server_opts, blank_database_server_opts) {
		// Shouldn't ever get here
		return fmt.Errorf("[ERROR] No database update changes detected for %s", d.Id())
	}
	database_server_opts.Id = d.Id()
	database_server, err := updateDatabaseServer(client, database_server_opts)
	if err != nil {
		return err
	}
	setDatabaseServerAttributes(d, database_server)
	setAllowAccessAttribute(d, database_server)
	d.Partial(false)
	return nil
}

func resourceBrightboxDatabaseServerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient
	log.Printf("[DEBUG] Database Server read called for %s", d.Id())
	database_server, err := client.DatabaseServer(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Database Server not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Database Server details: %s", err)
	}
	setDatabaseServerAttributes(d, database_server)
	setAllowAccessAttribute(d, database_server)
	return nil
}

func resourceBrightboxDatabaseServerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Database Server delete called for %s", d.Id())
	err := client.DestroyDatabaseServer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Database Server: %s", err)
	}
	stateConf := resource.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    databaseServerStateRefresh(client, d.Id()),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
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
	return nil
}

func output_database_server_options(opts *brightbox.DatabaseServerOptions) {
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
