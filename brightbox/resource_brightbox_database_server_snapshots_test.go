package brightbox

import (
	"context"
	"testing"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gotest.tools/v3/assert"
)

// TestUnitBrightboxDatabaseServer_SnapshotsScheduleCanBeExplicitlyCleared
// checks that an explicit `snapshots_schedule = ""` reaches the update
// options as a clear, even when prior state holds a non-empty
// API-assigned value. Schema-level only; no API access required.
func TestUnitBrightboxDatabaseServer_SnapshotsScheduleCanBeExplicitlyCleared(t *testing.T) {
	r := resourceBrightboxDatabaseServer()

	// Prior state as if the API had assigned a default schedule on create.
	state := &terraform.InstanceState{
		ID: "dbs-testt",
		Attributes: map[string]string{
			"name":                "db server",
			"snapshots_schedule":  "0 8 * * *",
			"snapshots_retention": "",
			"allow_access.#":      "0",
		},
		// RawConfig is what suppressUnconfiguredSnapshotField reads via
		// GetRawConfig; real applies populate it from the plugin protocol,
		// so it's set explicitly here.
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"name":                cty.StringVal("db server"),
			"snapshots_schedule":  cty.StringVal(""),
			"snapshots_retention": cty.NullVal(cty.String),
			"allow_access":        cty.SetValEmpty(cty.String),
		}),
	}

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"name":               "db server",
		"snapshots_schedule": "",
		"allow_access":       []interface{}{},
	})

	diff, err := r.Diff(context.Background(), state, config, nil)
	if err != nil {
		t.Fatalf("unexpected error computing diff: %s", err)
	}
	if diff == nil {
		t.Fatal("expected a diff clearing snapshots_schedule, got nil")
	}
	attrDiff, ok := diff.Attributes["snapshots_schedule"]
	if !ok {
		t.Fatal("expected snapshots_schedule to appear in the diff when explicitly cleared in config, but it was absent")
	}
	if attrDiff.New != "" {
		t.Errorf("expected snapshots_schedule to diff to \"\", got %q", attrDiff.New)
	}

	d, err := schema.InternalMap(r.Schema).Data(state, diff)
	if err != nil {
		t.Fatalf("unexpected error building resource data: %s", err)
	}

	var opts brightbox.DatabaseServerOptions
	if diags := addUpdateableDatabaseServerOptions(d, &opts); diags.HasError() {
		t.Fatalf("unexpected error: %#v", diags)
	}
	if opts.SnapshotsSchedule == nil {
		t.Fatal("expected snapshots_schedule to be sent to the API to clear it, but it was omitted from the update")
	}
	if *opts.SnapshotsSchedule != "" {
		t.Errorf("expected snapshots_schedule to be cleared to \"\", got %q", *opts.SnapshotsSchedule)
	}
	if !snapshotsScheduleCleared(d) {
		t.Error("expected snapshotsScheduleCleared to report the field as cleared")
	}
}

// TestUnitBrightboxDatabaseServer_SnapshotsScheduleSurvivesPlanAndApply
// drives resourceBrightboxDatabaseServer through a full Plan -> Apply cty
// cycle, not just the plan-time diff. A Computed attribute forced to
// New == "" is promoted to unknown ("computed on apply"), and unknown
// values are stripped before Apply rebuilds its diff - so a clear driven
// only by CustomizeDiff on a Computed attribute would vanish before
// Update ever saw it. suppressUnconfiguredSnapshotField avoids this by
// keeping the attribute Optional-only; this test is what would catch a
// regression back to Computed.
func TestUnitBrightboxDatabaseServer_SnapshotsScheduleSurvivesPlanAndApply(t *testing.T) {
	r := resourceBrightboxDatabaseServer()
	block := r.CoreConfigSchema()
	ty := block.ImpliedType()
	timeoutsTy := ty.AttributeType("timeouts")

	// Prior state as if the API had assigned a default schedule on create.
	priorVal := fullSchemaObject(ty, map[string]cty.Value{
		"id":                  cty.StringVal("dbs-testt"),
		"name":                cty.StringVal("db server"),
		"description":         cty.StringVal("db server"),
		"database_engine":     cty.StringVal("mysql"),
		"database_version":    cty.StringVal("8.0"),
		"database_type":       cty.StringVal("dbt-abcde"),
		"snapshots_schedule":  cty.StringVal("0 8 * * *"),
		"snapshots_retention": cty.StringVal(""),
		"allow_access":        cty.SetValEmpty(cty.String),
		"maintenance_weekday": cty.NumberIntVal(6),
		"maintenance_hour":    cty.NumberIntVal(6),
		"locked":              cty.False,
		"status":              cty.StringVal("active"),
		"zone":                cty.StringVal("zon-abcde"),
		"timeouts":            cty.NullVal(timeoutsTy),
	})

	// Config explicitly clears snapshots_schedule.
	configVal := fullSchemaObject(ty, map[string]cty.Value{
		"name":                cty.StringVal("db server"),
		"description":         cty.StringVal("db server"),
		"database_engine":     cty.StringVal("mysql"),
		"database_version":    cty.StringVal("8.0"),
		"database_type":       cty.StringVal("dbt-abcde"),
		"snapshots_schedule":  cty.StringVal(""),
		"allow_access":        cty.SetValEmpty(cty.String),
		"maintenance_weekday": cty.NumberIntVal(6),
		"maintenance_hour":    cty.NumberIntVal(6),
		"zone":                cty.StringVal("zon-abcde"),
		"timeouts":            cty.NullVal(timeoutsTy),
	})

	priorState, err := r.ShimInstanceStateFromValue(priorVal)
	if err != nil {
		t.Fatalf("shim prior state: %s", err)
	}
	priorState.RawConfig = configVal

	// --- Plan, exactly as PlanResourceChange would. ---
	planDiff, err := r.Diff(context.Background(), priorState, terraform.NewResourceConfigShimmed(configVal, block), nil)
	if err != nil {
		t.Fatalf("plan diff: %s", err)
	}

	plannedVal, err := schema.ApplyDiff(priorVal, planDiff, block)
	if err != nil {
		t.Fatalf("apply plan diff to prior value: %s", err)
	}
	if !plannedVal.GetAttr("snapshots_schedule").IsKnown() {
		t.Fatalf("planned snapshots_schedule is unknown (will be discarded before apply) instead of a concrete cleared value: %#v", plannedVal.GetAttr("snapshots_schedule"))
	}

	// --- Apply, exactly as ApplyResourceChange would: the diff is
	// reconstructed from prior/planned/config values with CustomizeDiff
	// stripped (resourceBrightboxDatabaseServer registers none, so r is
	// already equivalent to the stripped copy the real provider server
	// would use). ---
	applyDiff, err := schema.DiffFromValues(context.Background(), priorVal, plannedVal, configVal, r)
	if err != nil {
		t.Fatalf("apply-time diff: %s", err)
	}

	d, err := schema.InternalMap(r.Schema).Data(priorState, applyDiff)
	if err != nil {
		t.Fatalf("unexpected error building resource data: %s", err)
	}

	var opts brightbox.DatabaseServerOptions
	if diags := addUpdateableDatabaseServerOptions(d, &opts); diags.HasError() {
		t.Fatalf("unexpected error: %#v", diags)
	}
	if opts.SnapshotsSchedule == nil {
		t.Fatal("expected snapshots_schedule to survive Plan->Apply and be sent to the API to clear it, but it was omitted from the update")
	}
	if *opts.SnapshotsSchedule != "" {
		t.Errorf("expected snapshots_schedule to be cleared to \"\", got %q", *opts.SnapshotsSchedule)
	}
}

func TestRequestBodyWithNullFieldClearsBothDatabaseServerSnapshotFields(t *testing.T) {
	databaseServerID := "dbs-12345"
	name := "test-database-server"
	retention := ""
	schedule := ""
	databaseServerOpts := brightbox.DatabaseServerOptions{
		ID:                 databaseServerID,
		Name:               &name,
		SnapshotsRetention: &retention,
		SnapshotsSchedule:  &schedule,
	}

	requestBody, err := requestBodyWithNullField(databaseServerOpts, "snapshots_retention", "snapshots_schedule")

	assert.NilError(t, err)
	assert.Equal(t, requestBody["name"], name)
	assert.Equal(t, requestBody["snapshots_retention"], nil)
	assert.Equal(t, requestBody["snapshots_schedule"], nil)
	_, ok := requestBody["ID"]
	assert.Equal(t, ok, false)
}

func TestUnitBrightboxDatabaseServer_AddUpdateableOptionsOmitsUnsetSnapshotsSchedule(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceBrightboxDatabaseServer().Schema, map[string]interface{}{"allow_access": []interface{}{"any"}})
	var databaseServerOpts brightbox.DatabaseServerOptions
	assert.Assert(t, !addUpdateableDatabaseServerOptions(resourceData, &databaseServerOpts).HasError())
	assert.Assert(t, databaseServerOpts.SnapshotsSchedule == nil, "expected SnapshotsSchedule to be omitted when unset in config")
}

func TestUnitBrightboxDatabaseServer_AddUpdateableOptionsIncludesSetSnapshotsSchedule(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceBrightboxDatabaseServer().Schema, map[string]interface{}{
		"allow_access":       []interface{}{"any"},
		"snapshots_schedule": "0 17 * * *",
	})
	var databaseServerOpts brightbox.DatabaseServerOptions
	assert.Assert(t, !addUpdateableDatabaseServerOptions(resourceData, &databaseServerOpts).HasError())
	assert.Assert(t, databaseServerOpts.SnapshotsSchedule != nil, "expected SnapshotsSchedule to be set")
	assert.Equal(t, *databaseServerOpts.SnapshotsSchedule, "0 17 * * *")
}
