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
	diags := resourceBrightboxFirewallPolicyCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}
	return assignFirewallPolicy(ctx, d, meta)
}

func assignFirewallPolicy(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	if targetID, ok := d.GetOk("server_group"); ok {
		log.Printf("[INFO] Attaching %s to %s", d.Id(), targetID.(string))
		client := meta.(*CompositeClient).APIClient
		FirewallPolicyInstance, err := client.ApplyFirewallPolicy(
			ctx,
			d.Id(),
			brightbox.FirewallPolicyAttachment{targetID.(string)},
		)
		if err != nil {
			return diag.FromErr(err)
		}
		return setFirewallPolicyAttributes(d, FirewallPolicyInstance)
	}
	return nil
}

func unassignFirewallPolicy(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient
	targetID, _ := d.GetChange("server_group")
	if targetID.(string) != "" {
		log.Printf("[INFO] Detaching %s from %s", d.Id(), targetID.(string))
		_, err := client.RemoveFirewallPolicy(
			ctx,
			d.Id(),
			brightbox.FirewallPolicyAttachment{targetID.(string)},
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
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
		diags = append(diags, assignFirewallPolicy(ctx, d, meta)...)
	}
	return append(diags, resourceBrightboxFirewallPolicyUpdate(ctx, d, meta)...)
}

func firewallPolicyFromID(
	id string,
) *brightbox.FirewallPolicyOptions {
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
