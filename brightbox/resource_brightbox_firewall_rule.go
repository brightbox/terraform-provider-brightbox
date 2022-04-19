package brightbox

import (
	"context"
	"log"
	"regexp"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	validFirewallRuleProtocols = []string{"tcp", "udp", "icmp"}
	validICMPTypeRegexp        = regexp.MustCompile("^[a-z-]+$")
	validMultiplePortRegexp    = regexp.MustCompile("^[0-9:,-]+$")
)

func resourceBrightboxFirewallRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Firewall Rule resource",
		CreateContext: resourceBrightboxFirewallRuleCreate,
		ReadContext: resourceBrightboxRead(
			(*brightbox.Client).FirewallRule,
			"Firewall Rule",
			setFirewallRuleAttributes,
		),
		UpdateContext: resourceBrightboxFirewallRuleUpdateAndRegenerate,
		DeleteContext: resourceBrightboxFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"description": {
				Description: "User editable label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"destination": {
				Description:   "Subnet, ServerGroup or ServerID",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source"},
				ValidateFunc:  stringIsValidFirewallTarget(),
			},

			"destination_port": {
				Description: "single port, multiple ports or range separated by '-' or ':'; upto 255 characters example - '80', '80,443,21' or '3000-3999'",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(validMultiplePortRegexp, "must be a valid set of port ranges"),
				),
			},

			"firewall_policy": {
				Description:  "The firewall policy this rule is linked to",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(firewallPolicyRegexp, "must be a valid firewall policy ID"),
			},

			"icmp_type_name": {
				Description: "ICMP type name. 'echo-request', 'echo-reply'",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.Any(
					validation.IntBetween(0, 255),
					validation.StringMatch(validICMPTypeRegexp, "must be a valid ICMP type"),
				),
			},

			"protocol": {
				Description: "Protocol Number, or one of tcp, udp, icmp",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.Any(
					validation.StringInSlice(
						validFirewallRuleProtocols,
						false,
					),
					validation.IntBetween(0, 255),
				),
			},

			"source": {
				Description:   "Subnet, ServerGroup or ServerID",
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  stringIsValidFirewallTarget(),
				ConflictsWith: []string{"destination"},
			},

			"source_port": {
				Description: "single port, multiple ports or range separated by '-' or ':'; upto 255 characters example - '80', '80,443,21' or '3000-3999'",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(validMultiplePortRegexp, "must be a valid set of port ranges"),
				),
			},
		},
	}
}

var resourceBrightboxFirewallRuleCreate = resourceBrightboxCreate(
	(*brightbox.Client).CreateFirewallRule,
	"Firewall Rule",
	addUpdateableFirewallRuleOptions,
	setFirewallRuleAttributes,
)

var resourceBrightboxFirewallRuleDelete = resourceBrightboxDelete(
	(*brightbox.Client).DestroyFirewallRule,
	"Firewall Rule",
)

var resourceBrightboxFirewallRuleUpdate = resourceBrightboxUpdate(
	(*brightbox.Client).UpdateFirewallRule,
	"Firewall Rule",
	firewallRuleFromID,
	addUpdateableFirewallRuleOptions,
	setFirewallRuleAttributes,
)

func resourceBrightboxFirewallRuleUpdateAndRegenerate(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {

	if d.HasChange("firewall_policy") {
		log.Printf("[INFO] Firewall Policy changed, regenerating rule")
		log.Printf("[INFO] Removing original rule %s", d.Id())

		errs := resourceBrightboxFirewallRuleDelete(ctx, d, meta)
		if errs.HasError() {
			return errs
		}

		log.Printf("[INFO] Creating new rule")
		return resourceBrightboxFirewallRuleCreate(ctx, d, meta)
	}
	return resourceBrightboxFirewallRuleUpdate(ctx, d, meta)
}

func firewallRuleFromID(
	id string,
) *brightbox.FirewallRuleOptions {
	return &brightbox.FirewallRuleOptions{
		ID: id,
	}
}

func addUpdateableFirewallRuleOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallRuleOptions,
) diag.Diagnostics {
	if d.HasChange("firewall_policy") {
		opts.FirewallPolicy = d.Get("firewall_policy").(string)
	}
	if got, ok := d.GetOk("protocol"); ok {
		temp := got.(string)
		opts.Protocol = &temp
	}
	if got, ok := d.GetOk("source"); ok {
		temp := got.(string)
		opts.Source = &temp
	}
	if got, ok := d.GetOk("source_port"); ok {
		temp := got.(string)
		opts.SourcePort = &temp
	}
	if got, ok := d.GetOk("destination"); ok {
		temp := got.(string)
		opts.Destination = &temp
	}
	if got, ok := d.GetOk("destination_port"); ok {
		temp := got.(string)
		opts.DestinationPort = &temp
	}
	if got, ok := d.GetOk("icmp_type_name"); ok {
		temp := got.(string)
		opts.IcmpTypeName = &temp
	}
	if got, ok := d.GetOk("description"); ok {
		temp := got.(string)
		opts.Description = &temp
	} else {
		// Check if description has been cleared
		assignString(d, &opts.Description, "description")
	}
	return nil
}

func setFirewallRuleAttributes(
	d *schema.ResourceData,
	firewallRule *brightbox.FirewallRule,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(firewallRule.ID)
	err = d.Set("firewall_policy", firewallRule.FirewallPolicy.ID)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("protocol", firewallRule.Protocol)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source", firewallRule.Source)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("source_port", firewallRule.SourcePort)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("destination", firewallRule.Destination)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("destination_port", firewallRule.DestinationPort)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("icmp_type_name", firewallRule.IcmpTypeName)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("description", firewallRule.Description)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	return diags
}
