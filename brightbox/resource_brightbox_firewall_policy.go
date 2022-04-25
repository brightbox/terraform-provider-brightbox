package brightbox

import (
	"context"
	"errors"
	"log"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Firewall Policy resource",
		CreateContext: resourceBrightboxFirewallPolicyCreateAndAssign,
		ReadContext: resourceBrightboxRead(
			(*brightbox.Client).FirewallPolicy,
			"Firewall Policy",
			setFirewallPolicyAttributes,
		),
		UpdateContext: resourceBrightboxFirewallPolicyUpdateAndRemap,
		DeleteContext: resourceBrightboxDelete(
			(*brightbox.Client).DestroyFirewallPolicy,
			"Firewall Policy",
		),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"description": {
				Description: "Optional description of the policy",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"name": {
				Description: "Optional name for this policy",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"server_group": {
				Description:  "The server group using this policy",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(serverGroupRegexp, "must be a valid server group ID"),
			},
		},
	}
}

var resourceBrightboxFirewallPolicyCreate = resourceBrightboxCreate(
	(*brightbox.Client).CreateFirewallPolicy,
	"Firewall Policy",
	addUpdateableFirewallPolicyOptions,
	setFirewallPolicyAttributes,
)

func resourceBrightboxFirewallPolicyCreateAndAssign(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	targetID, ok := d.GetOk("server_group")
	diags := resourceBrightboxFirewallPolicyCreate(ctx, d, meta)
	if !ok || diags.HasError() {
		return diags
	}
	FirewallPolicyInstance, err := assignFirewallPolicy(ctx, d, meta, targetID.(string))
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] setting details from returned object")
	return setFirewallPolicyAttributes(d, FirewallPolicyInstance)
}

func assignFirewallPolicy(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
	target string,
) (*brightbox.FirewallPolicy, error) {
	log.Printf("[INFO] Attaching %s to %s", d.Id(), target)
	client := meta.(*CompositeClient).APIClient
	return client.ApplyFirewallPolicy(
		ctx,
		d.Id(),
		brightbox.FirewallPolicyAttachment{ServerGroup: target},
	)
}

func unassignFirewallPolicy(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient
	targetID, _ := d.GetChange("server_group")
	if target := targetID.(string); target != "" {
		log.Printf("[INFO] Detaching %s from %s", d.Id(), target)
		_, err := client.RemoveFirewallPolicy(
			ctx,
			d.Id(),
			brightbox.FirewallPolicyAttachment{ServerGroup: target},
		)
		if err == nil {
			log.Printf("[DEBUG] detached cleanly")
			return nil
		}
		log.Printf("[DEBUG] detachment failed - checking for out of band detachment")
		instance, readerr := client.FirewallPolicy(ctx, d.Id())
		if readerr != nil {
			return diag.FromErr(readerr)
		}
		if !detachedFirewallPolicy(instance) {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] detached out of band")
	}
	return nil
}

func detachedFirewallPolicy(instance *brightbox.FirewallPolicy) bool {
	return instance.ServerGroup == nil
}

var resourceBrightboxFirewallPolicyUpdate = resourceBrightboxUpdate(
	(*brightbox.Client).UpdateFirewallPolicy,
	"Firewall Policy",
	firewallPolicyFromID,
	addUpdateableFirewallPolicyOptions,
	setFirewallPolicyAttributes,
)

func resourceBrightboxFirewallPolicyUpdateAndRemap(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	var diags diag.Diagnostics

	if d.HasChange("server_group") {
		log.Printf("[INFO] Server Group changed, updating...")
		diags = append(diags, unassignFirewallPolicy(ctx, d, meta)...)
		if targetID, ok := d.GetOk("server_group"); ok {
			if target := targetID.(string); target != "" {
				_, err := assignFirewallPolicy(ctx, d, meta, target)
				if err != nil {
					diags = append(diags, diag.FromErr(err)...)
				}
			}
		}
	}
	return append(diags, resourceBrightboxFirewallPolicyUpdate(ctx, d, meta)...)
}

func firewallPolicyFromID(id string) *brightbox.FirewallPolicyOptions {
	return &brightbox.FirewallPolicyOptions{
		ID: id,
	}
}

func addUpdateableFirewallPolicyOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallPolicyOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Description, "description")
	return nil
}

func setFirewallPolicyAttributes(
	d *schema.ResourceData,
	firewallPolicy *brightbox.FirewallPolicy,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(firewallPolicy.ID)
	err = d.Set("name", firewallPolicy.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", firewallPolicy.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	if firewallPolicy.ServerGroup == nil {
		err = d.Set("server_group", "")
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	} else {
		err = d.Set("server_group", firewallPolicy.ServerGroup.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	return diags
}

func retryServerGroupChange(changeFunc func() error, timeout time.Duration) error {
	// Wait for group to change
	return resource.Retry(
		timeout,
		func() *resource.RetryError {
			if err := changeFunc(); err != nil {
				var apierror *brightbox.APIError
				if errors.As(err, &apierror) {
					if apierror.StatusCode == 409 {
						return resource.RetryableError(err)
					}
				}
				return resource.NonRetryableError(err)
			}
			return nil
		},
	)
}
