package brightbox

import (
	"context"
	"log"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/databaseserverstatus"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxDatabaseServer() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Database Server resource",
		CreateContext: resourceBrightboxDatabaseServerCreateAndWait,
		ReadContext:   resourceBrightboxDatabaseServerRead,
		UpdateContext: resourceBrightboxDatabaseServerResizeAndUpdate,
		DeleteContext: resourceBrightboxDatabaseServerDeleteAndWait,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"admin_password": {
				Description: "Initial password required to login, only available at creation or following a password reset request",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},

			"admin_username": {
				Description: "Initial username required to login",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"allow_access": {
				Description: "An array of resources allowed to access the database. Accepted values include `any`, `IPv4 address`, `server identifier`, `server group identifier`",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: stringIsValidFirewallTarget(),
				},
				Required: true,
				MinItems: 1,
				Set:      schema.HashString,
			},

			"database_engine": {
				Description: "The DBMS engine of the Database Server",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					validDatabaseEngines,
					false,
				),
			},

			"database_type": {
				Description:  "ID of the database type to use",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(databaseTypeRegexp, "must be a valid database type ID"),
			},

			"database_version": {
				Description:  "The version of the given engine in use",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"description": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"locked": {
				Description: "Initial password required to login, only available at creation or following a password reset request",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"maintenance_hour": {
				Description:  "Number representing 24hr time start of maintenance window hour for x:00-x:59 (0-23)",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 23),
			},

			"maintenance_weekday": {
				Description:  "Numerical index of weekday (0 is Sunday, 1 is Monday...) to set when automatic updates may be performed",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 6),
			},

			"name": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"snapshot": {
				Description:  "Identifier for an SQL snapshot to use as the basis of the new instance. Creates and restores the database from the snapshot",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(databaseSnapshotRegexp, "must be a valid database snapshot ID"),
			},

			"snapshots_retention": {
				Description:  "Keep this number of scheduled snapshots. Keep all if unset",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},

			"snapshots_schedule": {
				Description:  "Crontab pattern for scheduled snapshots. Must be at least hourly",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: ValidateCronString,
			},

			"snapshots_schedule_next_at": {
				Description: "time in UTC when next approximate scheduled snapshot will be run",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"status": {
				Description: "State the database server is in",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"zone": {
				Description:  "ID of the zone the database server is in",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(zoneRegexp, "must be a valid zone ID or handle"),
			},
		},
	}
}

var (
	resourceBrightboxDatabaseServerRead = resourceBrightboxReadStatus(
		(*brightbox.Client).DatabaseServer,
		"Load Balancer",
		setDatabaseServerAttributes,
		databaseServerUnavailable,
	)

	resourceBrightboxDatabaseServerUpdate = resourceBrightboxUpdateWithLock(
		(*brightbox.Client).UpdateDatabaseServer,
		"Database Server",
		databaseServerFromID,
		addUpdateableDatabaseServerOptions,
		setDatabaseServerAttributes,
		resourceBrightboxSetDatabaseServerLockState,
	)

	resourceBrightboxDatabaseServerDeleteAndWait = resourceBrightboxDeleteAndWait(
		(*brightbox.Client).DestroyDatabaseServer,
		"Database Server",
		[]string{
			databaseserverstatus.Deleting.String(),
			databaseserverstatus.Active.String(),
		},
		[]string{
			databaseserverstatus.Deleted.String(),
		},
		databaseServerStateRefresh,
	)

	resourceBrightboxSetDatabaseServerLockState = resourceBrightboxSetLockState(
		(*brightbox.Client).LockDatabaseServer,
		(*brightbox.Client).UnlockDatabaseServer,
		setDatabaseServerAttributes,
	)
)

func databaseServerFromID(id string) *brightbox.DatabaseServerOptions {
	return &brightbox.DatabaseServerOptions{
		ID: id,
	}
}

func addUpdateableDatabaseServerOptions(
	d *schema.ResourceData,
	opts *brightbox.DatabaseServerOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	assignByte(d, &opts.MaintenanceWeekday, "maintenance_weekday")
	assignByte(d, &opts.MaintenanceHour, "maintenance_hour")
	// Always set snapshot schedule to get around default issue
	schedule := d.Get("snapshots_schedule").(string)
	opts.SnapshotsSchedule = &schedule
	assignString(d, &opts.SnapshotsRetention, "snapshots_retention")
	assignStringSet(d, &opts.AllowAccess, "allow_access")
	return nil
}

func setDatabaseServerAttributes(
	d *schema.ResourceData,
	databaseServer *brightbox.DatabaseServer,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(databaseServer.ID)
	err = d.Set("name", databaseServer.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", databaseServer.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", databaseServer.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("locked", databaseServer.Locked)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("database_engine", databaseServer.DatabaseEngine)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("database_version", databaseServer.DatabaseVersion)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("database_type", databaseServer.DatabaseServerType.ID)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("admin_username", databaseServer.AdminUsername)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("maintenance_weekday", databaseServer.MaintenanceWeekday)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("maintenance_hour", databaseServer.MaintenanceHour)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("snapshots_retention", databaseServer.SnapshotsRetention)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("snapshots_schedule", databaseServer.SnapshotsSchedule)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	var snapshotTime string
	if databaseServer.SnapshotsScheduleNextAt != nil {
		snapshotTime = databaseServer.SnapshotsScheduleNextAt.Format(time.RFC3339)
	}
	err = d.Set("snapshots_schedule_next_at", snapshotTime)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("zone", databaseServer.Zone.Handle)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("allow_access", databaseServer.AllowAccess)
	if err != nil {
		diags = append(diags, diag.Errorf("error setting allow_access: %s", err)...)
	}
	return diags
}

func databaseServerUnavailable(obj *brightbox.DatabaseServer) bool {
	return obj.Status == databaseserverstatus.Deleted ||
		obj.Status == databaseserverstatus.Failed
}

func databaseServerStateRefresh(client *brightbox.Client, ctx context.Context, databaseServerID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		databaseServer, err := client.DatabaseServer(ctx, databaseServerID)
		if err != nil {
			log.Printf("Error on Database Server State Refresh: %s", err)
			return nil, "", err
		}
		return databaseServer, databaseServer.Status.String(), nil
	}
}

func resourceBrightboxDatabaseServerResizeAndUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Database Resize and Update called for %s", d.Id())
	if d.HasChange("database_type") {
		newDatabaseType := d.Get("database_type").(string)
		log.Printf("[INFO] Changing database type to %v", newDatabaseType)
		_, err := client.ResizeDatabaseServer(
			ctx,
			d.Id(),
			brightbox.DatabaseServerNewSize{NewType: newDatabaseType},
		)
		if err != nil {
			return brightboxFromErrSlice(err)
		}
	}
	return resourceBrightboxDatabaseServerUpdate(ctx, d, meta)
}

func resourceBrightboxDatabaseServerCreateAndWait(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO]] Creating Database Server")
	var databaseServerOpts brightbox.DatabaseServerOptions

	errs := addUpdateableDatabaseServerOptions(d, &databaseServerOpts)
	if errs.HasError() {
		return errs
	}

	engine := &databaseServerOpts.Engine
	assignString(d, &engine, "database_engine")
	version := &databaseServerOpts.Version
	assignString(d, &version, "database_version")
	databaseType := &databaseServerOpts.DatabaseType
	assignString(d, &databaseType, "database_type")
	snapshot := &databaseServerOpts.Snapshot
	assignString(d, &snapshot, "snapshot")
	zone := &databaseServerOpts.Zone
	assignString(d, &zone, "zone")

	log.Printf("[DEBUG] Database Server create configuration %+v", databaseServerOpts)
	outputDatabaseServerOptions(&databaseServerOpts)

	databaseServer, err := client.CreateDatabaseServer(ctx, databaseServerOpts)
	if err != nil {
		return brightboxFromErrSlice(err)
	}

	d.SetId(databaseServer.ID)

	if databaseServer.AdminPassword == "" {
		log.Printf("[WARN] No password returned for Cloud SQL server %s", databaseServer.ID)
	} else {
		d.Set("admin_password", databaseServer.AdminPassword)
	}

	log.Printf("[INFO] Waiting for Database Server (%s) to become available", d.Id())

	stateConf := retry.StateChangeConf{
		Pending: []string{
			databaseserverstatus.Creating.String(),
		},
		Target: []string{
			databaseserverstatus.Active.String(),
		},
		Refresh:    databaseServerStateRefresh(client, ctx, databaseServer.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return brightboxFromErrSlice(err)
	}

	return resourceBrightboxSetDatabaseServerLockState(ctx, d, meta)
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
		log.Printf("[DEBUG] Database Server Snapshot %q", opts.Snapshot)
	}
	if opts.SnapshotsRetention != nil {
		log.Printf("[DEBUG] Database Server Snapshots Retention %q", *opts.SnapshotsRetention)
	}
	if opts.SnapshotsSchedule != nil {
		log.Printf("[DEBUG] Database Server Snapshots Schedule %q", *opts.SnapshotsSchedule)
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
