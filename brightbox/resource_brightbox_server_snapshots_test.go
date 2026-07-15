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

// TestUnitBrightboxServer_SnapshotsScheduleCanBeExplicitlyCleared is the
// resource_brightbox_server counterpart to
// TestUnitBrightboxDatabaseServer_SnapshotsScheduleCanBeExplicitlyCleared
// (resource_brightbox_database_server_snapshots_test.go).
func TestUnitBrightboxServer_SnapshotsScheduleCanBeExplicitlyCleared(t *testing.T) {
	r := resourceBrightboxServer()

	state := &terraform.InstanceState{
		ID: "srv-testt",
		Attributes: map[string]string{
			"name":                "srv server",
			"snapshots_schedule":  "0 8 * * *",
			"snapshots_retention": "",
			"server_groups.#":     "0",
		},
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"name":                cty.StringVal("srv server"),
			"snapshots_schedule":  cty.StringVal(""),
			"snapshots_retention": cty.NullVal(cty.String),
			"server_groups":       cty.SetValEmpty(cty.String),
		}),
	}

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"name":               "srv server",
		"snapshots_schedule": "",
		"server_groups":      []interface{}{},
	})

	diff, err := r.Diff(context.Background(), state, config, nil)
	if err != nil {
		t.Fatalf("unexpected error computing diff: %s", err)
	}
	if diff == nil {
		t.Fatal("expected a diff clearing snapshots_schedule, got nil")
	}
	if _, ok := diff.Attributes["snapshots_schedule"]; !ok {
		t.Fatal("expected snapshots_schedule to appear in the diff when explicitly cleared in config, but it was absent")
	}

	d, err := schema.InternalMap(r.Schema).Data(state, diff)
	if err != nil {
		t.Fatalf("unexpected error building resource data: %s", err)
	}

	var opts brightbox.ServerOptions
	if diags := addUpdateableServerOptions(d, &opts); diags.HasError() {
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

// TestUnitBrightboxServer_SnapshotsScheduleSurvivesPlanAndApply is the
// resource_brightbox_server counterpart to
// TestUnitBrightboxDatabaseServer_SnapshotsScheduleSurvivesPlanAndApply
// (resource_brightbox_database_server_snapshots_test.go). fullSchemaObject
// is a shared helper in resource_brightbox_context_funcs_test.go.
func TestUnitBrightboxServer_SnapshotsScheduleSurvivesPlanAndApply(t *testing.T) {
	r := resourceBrightboxServer()
	block := r.CoreConfigSchema()
	ty := block.ImpliedType()
	timeoutsTy := ty.AttributeType("timeouts")

	// Prior state as if the API had assigned a default schedule on create.
	priorVal := fullSchemaObject(ty, map[string]cty.Value{
		"id":                  cty.StringVal("srv-testt"),
		"name":                cty.StringVal("srv server"),
		"image":               cty.StringVal("img-abcde"),
		"type":                cty.StringVal("1gb.ssd"),
		"snapshots_schedule":  cty.StringVal("0 8 * * *"),
		"snapshots_retention": cty.StringVal(""),
		"server_groups":       cty.SetValEmpty(cty.String),
		"locked":              cty.False,
		"status":              cty.StringVal("active"),
		"zone":                cty.StringVal("zon-abcde"),
		"timeouts":            cty.NullVal(timeoutsTy),
	})

	// Config explicitly clears snapshots_schedule.
	configVal := fullSchemaObject(ty, map[string]cty.Value{
		"name":               cty.StringVal("srv server"),
		"image":              cty.StringVal("img-abcde"),
		"type":               cty.StringVal("1gb.ssd"),
		"snapshots_schedule": cty.StringVal(""),
		"server_groups":      cty.SetValEmpty(cty.String),
		"zone":               cty.StringVal("zon-abcde"),
		"timeouts":           cty.NullVal(timeoutsTy),
	})

	priorState, err := r.ShimInstanceStateFromValue(priorVal)
	if err != nil {
		t.Fatalf("shim prior state: %s", err)
	}
	priorState.RawConfig = configVal

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

	applyDiff, err := schema.DiffFromValues(context.Background(), priorVal, plannedVal, configVal, r)
	if err != nil {
		t.Fatalf("apply-time diff: %s", err)
	}

	d, err := schema.InternalMap(r.Schema).Data(priorState, applyDiff)
	if err != nil {
		t.Fatalf("unexpected error building resource data: %s", err)
	}

	var opts brightbox.ServerOptions
	if diags := addUpdateableServerOptions(d, &opts); diags.HasError() {
		t.Fatalf("unexpected error: %#v", diags)
	}
	if opts.SnapshotsSchedule == nil {
		t.Fatal("expected snapshots_schedule to survive Plan->Apply and be sent to the API to clear it, but it was omitted from the update")
	}
	if *opts.SnapshotsSchedule != "" {
		t.Errorf("expected snapshots_schedule to be cleared to \"\", got %q", *opts.SnapshotsSchedule)
	}
}

// TestRequestBodyWithNullFieldClearsServerSnapshotFields is the
// resource_brightbox_server counterpart to
// TestRequestBodyWithNullFieldClearsBothDatabaseServerSnapshotFields
// (resource_brightbox_database_server_snapshots_test.go).
func TestRequestBodyWithNullFieldClearsServerSnapshotFields(t *testing.T) {
	serverID := "srv-12345"
	name := "test-server"
	retention := ""
	schedule := ""
	serverOpts := brightbox.ServerOptions{
		ID:                 serverID,
		Name:               &name,
		SnapshotsRetention: &retention,
		SnapshotsSchedule:  &schedule,
	}

	requestBody, err := requestBodyWithNullField(serverOpts, "snapshots_retention", "snapshots_schedule")

	assert.NilError(t, err)
	assert.Equal(t, requestBody["name"], name)
	assert.Equal(t, requestBody["snapshots_retention"], nil)
	assert.Equal(t, requestBody["snapshots_schedule"], nil)
	_, ok := requestBody["ID"]
	assert.Equal(t, ok, false)
}

func TestUnitBrightboxServer_AddUpdateableOptionsOmitsUnsetSnapshotsSchedule(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceBrightboxServer().Schema, map[string]interface{}{})
	var serverOpts brightbox.ServerOptions
	assert.Assert(t, !addUpdateableServerOptions(resourceData, &serverOpts).HasError())
	assert.Assert(t, serverOpts.SnapshotsSchedule == nil, "expected SnapshotsSchedule to be omitted when unset in config")
}

func TestUnitBrightboxServer_AddUpdateableOptionsIncludesSetSnapshotsSchedule(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceBrightboxServer().Schema, map[string]interface{}{
		"snapshots_schedule": "0 17 * * *",
	})
	var serverOpts brightbox.ServerOptions
	assert.Assert(t, !addUpdateableServerOptions(resourceData, &serverOpts).HasError())
	assert.Assert(t, serverOpts.SnapshotsSchedule != nil, "expected SnapshotsSchedule to be set")
	assert.Equal(t, *serverOpts.SnapshotsSchedule, "0 17 * * *")
}
