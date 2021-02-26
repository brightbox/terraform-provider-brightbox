package brightbox

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
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
		Description: "Provides a Brightbox Firewall Rule resource",
		Create:      resourceBrightboxFirewallRuleCreate,
		Read:        resourceBrightboxFirewallRuleRead,
		Update:      resourceBrightboxFirewallRuleUpdate,
		Delete:      resourceBrightboxFirewallRuleDelete,
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

func resourceBrightboxFirewallRuleCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Creating Firewall Rule")
	protocol := d.Get("protocol").(string)
	source := d.Get("source").(string)
	sourcePort := d.Get("source_port").(string)
	destination := d.Get("destination").(string)
	destinationPort := d.Get("destination_port").(string)
	icmpTypeName := d.Get("icmp_type_name").(string)
	description := d.Get("description").(string)
	firewallRuleOpts := &brightbox.FirewallRuleOptions{
		FirewallPolicy:  d.Get("firewall_policy").(string),
		Protocol:        &protocol,
		Source:          &source,
		SourcePort:      &sourcePort,
		Destination:     &destination,
		DestinationPort: &destinationPort,
		IcmpTypeName:    &icmpTypeName,
		Description:     &description,
	}

	log.Printf("[INFO] Firewall Rule create configuration: %#v", firewallRuleOpts)

	firewallRule, err := client.CreateFirewallRule(firewallRuleOpts)
	if err != nil {
		return fmt.Errorf("Error creating Firewall Rule: %s", err)
	}

	d.SetId(firewallRule.Id)

	return setFirewallRuleAttributes(d, firewallRule)
}

func resourceBrightboxFirewallRuleRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	firewallRule, err := client.FirewallRule(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Firewall Rule not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Firewall Rule details: %s", err)
	}

	return setFirewallRuleAttributes(d, firewallRule)
}

func resourceBrightboxFirewallRuleDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO] Deleting Firewall Rule %s", d.Id())
	err := client.DestroyFirewallRule(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Firewall Rule (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceBrightboxFirewallRuleUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	if d.HasChange("firewall_policy") {
		log.Printf("[INFO] Firewall Policy changed, regenerating rule")
		oldFirewallRuleID := d.Id()
		err := resourceBrightboxFirewallRuleCreate(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Removing original rule %s", oldFirewallRuleID)
		err = client.DestroyFirewallRule(oldFirewallRuleID)
		if err != nil {
			return fmt.Errorf("Error deleting Firewall Rule (%s): %s", oldFirewallRuleID, err)
		}
		return nil

	}
	firewallRuleOpts := &brightbox.FirewallRuleOptions{
		Id: d.Id(),
	}
	err := addUpdateableFirewallRuleOptions(d, firewallRuleOpts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Firewall Rule update configuration: %#v", firewallRuleOpts)
	firewallRule, err := client.UpdateFirewallRule(firewallRuleOpts)
	if err != nil {
		return fmt.Errorf("Error updating Firewall Rule (%s): %s", firewallRuleOpts.Id, err)
	}

	return setFirewallRuleAttributes(d, firewallRule)
}

func addUpdateableFirewallRuleOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallRuleOptions,
) error {
	assignString(d, &opts.Protocol, "protocol")
	assignString(d, &opts.Source, "source")
	assignString(d, &opts.SourcePort, "source_port")
	assignString(d, &opts.Destination, "destination")
	assignString(d, &opts.DestinationPort, "destination_port")
	assignString(d, &opts.IcmpTypeName, "icmp_type_name")
	assignString(d, &opts.Description, "description")
	return nil
}

func setFirewallRuleAttributes(
	d *schema.ResourceData,
	firewallRule *brightbox.FirewallRule,
) error {
	d.Set("firewall_policy", firewallRule.FirewallPolicy.Id)
	d.Set("protocol", firewallRule.Protocol)
	d.Set("source", firewallRule.Source)
	d.Set("source_port", firewallRule.SourcePort)
	d.Set("destination", firewallRule.Destination)
	d.Set("destination_port", firewallRule.DestinationPort)
	d.Set("icmp_type_name", firewallRule.IcmpTypeName)
	d.Set("description", firewallRule.Description)
	return nil
}
