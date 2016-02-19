package brightbox

import (
	"fmt"
	"log"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBrightboxFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxFirewallRuleCreate,
		Read:   resourceBrightboxFirewallRuleRead,
		Update: resourceBrightboxFirewallRuleUpdate,
		Delete: resourceBrightboxFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			"firewall_policy": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"source": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"source_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"destination": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"destination_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"icmp_type_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceBrightboxFirewallRuleCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

	log.Printf("[INFO] Creating Firewall Rule")
	firewall_rule_opts := &brightbox.FirewallRuleOptions{
		FirewallPolicy: d.Get("firewall_policy").(string),
	}
	err := addUpdateableFirewallRuleOptions(d, firewall_rule_opts)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Firewall Rule create configuration: %#v", firewall_rule_opts)

	firewall_rule, err := client.CreateFirewallRule(firewall_rule_opts)
	if err != nil {
		return fmt.Errorf("Error creating Firewall Rule: %s", err)
	}

	d.SetId(firewall_rule.Id)

	setFirewallRuleAttributes(d, firewall_rule)

	return nil
}

func setFirewallRuleAttributes(
	d *schema.ResourceData,
	firewall_rule *brightbox.FirewallRule,
) {
	d.Set("firewall_policy", firewall_rule.FirewallPolicy)
	d.Set("protocol", firewall_rule.Protocol)
	d.Set("source", firewall_rule.Source)
	d.Set("source_port", firewall_rule.SourcePort)
	d.Set("destination", firewall_rule.Destination)
	d.Set("destination_port", firewall_rule.DestinationPort)
	d.Set("icmp_type_name", firewall_rule.IcmpTypeName)
	d.Set("description", firewall_rule.Description)
}

func resourceBrightboxFirewallRuleRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

	firewall_rule, err := client.FirewallRule(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Firewall Rule details: %s", err)
	}

	setFirewallRuleAttributes(d, firewall_rule)

	return nil
}

func resourceBrightboxFirewallRuleDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*brightbox.Client)

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
	client := meta.(*brightbox.Client)

	firewall_rule_opts := &brightbox.FirewallRuleOptions{
		Id: d.Id(),
	}
	err := addUpdateableFirewallRuleOptions(d, firewall_rule_opts)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Firewall Rule update configuration: %#v", firewall_rule_opts)

	firewall_rule, err := client.UpdateFirewallRule(firewall_rule_opts)
	if err != nil {
		return fmt.Errorf("Error updating Firewall Rule (%s): %s", firewall_rule_opts.Id, err)
	}

	setFirewallRuleAttributes(d, firewall_rule)
	return nil
}

func addUpdateableFirewallRuleOptions(
	d *schema.ResourceData,
	opts *brightbox.FirewallRuleOptions,
) error {
	if d.HasChange("protocol") {
		var temp string
		if attr, ok := d.GetOk("protocol"); ok {
			temp = attr.(string)
		}
		opts.Protocol = &temp
	}
	if d.HasChange("source") {
		var temp string
		if attr, ok := d.GetOk("source"); ok {
			temp = attr.(string)
		}
		opts.Source = &temp
	}
	if d.HasChange("source_port") {
		var temp string
		if attr, ok := d.GetOk("source_port"); ok {
			temp = attr.(string)
		}
		opts.SourcePort = &temp
	}
	if d.HasChange("destination") {
		var temp string
		if attr, ok := d.GetOk("destination"); ok {
			temp = attr.(string)
		}
		opts.Destination = &temp
	}
	if d.HasChange("destination_port") {
		var temp string
		if attr, ok := d.GetOk("destination_port"); ok {
			temp = attr.(string)
		}
		opts.DestinationPort = &temp
	}
	if d.HasChange("icmp_type_name") {
		var temp string
		if attr, ok := d.GetOk("icmp_type_name"); ok {
			temp = attr.(string)
		}
		opts.IcmpTypeName = &temp
	}
	if d.HasChange("description") {
		var temp string
		if attr, ok := d.GetOk("description"); ok {
			temp = attr.(string)
		}
		opts.Description = &temp
	}
	return nil
}
