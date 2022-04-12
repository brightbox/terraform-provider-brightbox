package brightbox

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/status/cloudip"
	"github.com/brightbox/gobrightbox/v2/status/mode"
	"github.com/brightbox/gobrightbox/v2/status/transportprotocol"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBrightboxCloudip() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox CloudIP resource",
		CreateContext: resourceBrightboxCloudipCreateAndAssign,
		ReadContext: resourceBrightboxRead(
			(*brightbox.Client).CloudIP,
			"Cloud IP",
			setCloudipAttributes,
		),
		UpdateContext: resourceBrightboxCloudipUpdateAndRemap,
		DeleteContext: resourceBrightboxCloudipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"fqdn": {
				Description: "Full Domain name entry for the Cloud IP",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"locked": {
				Description: "No lock on Cloud IPs",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "No lock on Cloud IPs",
			},

			"mode": {
				Description: "Type of Cloud IP required (nat/route)",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.StringInSlice(
					mode.ValidStrings,
					false,
				),
			},

			"name": {
				Description: "Name assigned to the Cloud IP",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"port_translator": {
				Description: "Array of Port Translators",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"incoming": {
							Description:  "Incoming Port",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"outgoing": {
							Description:  "Outgoing Port",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"protocol": {
							Description: "Transport protocol to port translate (tcp/udp)",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								transportprotocol.ValidStrings,
								false,
							),
						},
					},
				},
				Set: resourceBrightboxPortTranslationHash,
			},

			"public_ip": {
				Description: "Old alias of the IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"public_ipv4": {
				Description: "IPv4 address",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"public_ipv6": {
				Description: "IPv6 address",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"reverse_dns": {
				Description:  "Reverse DNS entry for the Cloud IP",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(dnsNameRegexp, "must be a valid DNS name"),
			},

			"status": {
				Description: "Current state of the Cloud IP",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"target": {
				Description: "The object this Cloud IP maps to",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.Any(
					validation.StringMatch(interfaceRegexp, "must by a valid server interface ID"),
					validation.StringMatch(loadBalancerRegexp, "must be a valid load balancer ID"),
					validation.StringMatch(databaseServerRegexp, "must be a valid database server ID"),
					validation.StringMatch(serverGroupRegexp, "must be a valid serer group ID"),
				),
			},
		},
	}
}

var resourceBrightboxCloudipCreate = resourceBrightboxCreate(
	(*brightbox.Client).CreateCloudIP,
	"Cloud IP",
	addUpdateableCloudipOptions,
	setCloudipAttributes,
)

func resourceBrightboxCloudipCreateAndAssign(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	diags := resourceBrightboxCloudipCreate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	if targetID, ok := d.GetOk("target"); ok {
		client := meta.(*CompositeClient).APIClient
		cloudipInstance, err := assignCloudIP(ctx, client, d.Id(), targetID.(string), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)
			return diags
		}
		return setCloudipAttributes(d, cloudipInstance)
	}
	return diags
}

func resourceBrightboxCloudipDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient
	return removeCloudIP(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
}

var resourceBrightboxCloudipUpdate = resourceBrightboxUpdate(
	(*brightbox.Client).UpdateCloudIP,
	"Cloud IP",
	cloudIPFromID,
	addUpdateableCloudipOptions,
	setCloudipAttributes,
)

func resourceBrightboxCloudipUpdateAndRemap(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	if d.HasChange("target") {
		err := unmapCloudIP(ctx, client, d.Id(), d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
		if targetID, ok := d.GetOk("target"); ok {
			_, err := assignCloudIP(ctx, client, d.Id(), targetID.(string), d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceBrightboxCloudipUpdate(ctx, d, meta)
}

func assignCloudIP(
	ctx context.Context,
	client *brightbox.Client,
	cloudipID string,
	targetID string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	log.Printf("[INFO] Assigning Cloud IP %s to target %s", cloudipID, targetID)
	_, err := client.MapCloudIP(
		ctx,
		cloudipID,
		brightbox.CloudIPAttachment{targetID},
	)
	if err != nil {
		return nil, fmt.Errorf("Error assigning Cloud IP %s to target %s: %s", cloudipID, targetID, err)
	}
	cloudipInstance, err := waitForMappedCloudIP(ctx, client, cloudipID, timeout)
	if err != nil {
		return nil, err
	}
	return cloudipInstance, err
}

func unmapCloudIP(
	ctx context.Context,
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
) error {
	log.Printf("[INFO] Checking mapping of Cloud IP %s", cloudipID)
	cloudipInstance, err := client.CloudIP(ctx, cloudipID)
	if err != nil {
		return fmt.Errorf("Error retrieving details of Cloud IP %s: %s", cloudipID, err)
	}
	if cloudipInstance.Status == cloudip.Mapped {
		log.Printf("[INFO] Unmapping Cloud IP %s", cloudipID)
		_, err := client.UnMapCloudIP(ctx, cloudipID)
		if err != nil {
			return fmt.Errorf("Error unmapping Cloud IP %s: %s", cloudipID, err)
		}
		_, err = waitForUnmappedCloudIP(ctx, client, cloudipID, timeout)
		if err != nil {
			return err
		}
	} else {
		log.Printf("[DEBUG] Cloud IP %s is already unmapped", cloudipID)
	}
	return nil
}

func waitForCloudip(
	ctx context.Context,
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
	pending string,
	target string,
) (*brightbox.CloudIP, error) {
	stateConf := resource.StateChangeConf{
		Pending:    []string{pending},
		Target:     []string{target},
		Refresh:    cloudipStateRefresh(ctx, client, cloudipID),
		Timeout:    timeout,
		MinTimeout: minimumRefreshWait,
	}

	activeCloudIP, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	}

	return activeCloudIP.(*brightbox.CloudIP), err
}

func waitForMappedCloudIP(
	ctx context.Context,
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	return waitForCloudip(ctx, client, cloudipID, timeout,
		cloudip.Unmapped.String(),
		cloudip.Mapped.String(),
	)
}

func waitForUnmappedCloudIP(
	ctx context.Context,
	client *brightbox.Client,
	cloudipID string,
	timeout time.Duration,
) (*brightbox.CloudIP, error) {
	return waitForCloudip(ctx, client, cloudipID, timeout,
		cloudip.Mapped.String(),
		cloudip.Unmapped.String(),
	)
}

func cloudipStateRefresh(ctx context.Context, client *brightbox.Client, cloudipID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cloudipInstance, err := client.CloudIP(ctx, cloudipID)
		if err != nil {
			log.Printf("Error on Cloud IP State Refresh: %s", err)
			return nil, "", err
		}
		return cloudipInstance, cloudipInstance.Status.String(), nil
	}
}

func cloudIPFromID(
	id string,
) *brightbox.CloudIPOptions {
	return &brightbox.CloudIPOptions{
		ID: id,
	}
}

func setCloudipAttributes(
	d *schema.ResourceData,
	cloudipInstance *brightbox.CloudIP,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(cloudipInstance.ID)
	err = d.Set("name", cloudipInstance.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("public_ip", cloudipInstance.PublicIP)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("public_ipv4", cloudipInstance.PublicIPv4)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("public_ipv6", cloudipInstance.PublicIPv6)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", cloudipInstance.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("reverse_dns", cloudipInstance.ReverseDNS)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("fqdn", cloudipInstance.Fqdn)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	// Set the server id first and let interface override it
	// Server and interface should appear together, but catch at least one
	if cloudipInstance.Server != nil {
		err = d.Set("target", cloudipInstance.Server.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	if cloudipInstance.Interface != nil {
		err = d.Set("target", cloudipInstance.Interface.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	if cloudipInstance.LoadBalancer != nil {
		err = d.Set("target", cloudipInstance.LoadBalancer.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	if cloudipInstance.DatabaseServer != nil {
		err = d.Set("target", cloudipInstance.DatabaseServer.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	if cloudipInstance.ServerGroup != nil {
		err = d.Set("target", cloudipInstance.ServerGroup.ID)
		if err != nil {
			diags = append(diags, diag.Errorf("unexpected: %s", err)...)
		}
	}
	log.Printf("[DEBUG] PortTranslator details are %#v", cloudipInstance.PortTranslators)
	portTranslators := make([]map[string]interface{}, len(cloudipInstance.PortTranslators))
	for i, portTranslator := range cloudipInstance.PortTranslators {
		portTranslators[i] = map[string]interface{}{
			"incoming": portTranslator.Incoming,
			"outgoing": portTranslator.Outgoing,
			"protocol": portTranslator.Protocol.String(),
		}
	}
	if err := d.Set("port_translator", portTranslators); err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}

	return diags
}

func removeCloudIP(ctx context.Context, client *brightbox.Client, id string, timeout time.Duration) diag.Diagnostics {
	log.Printf("[DEBUG] Unmapping Cloud IP %s", id)
	err := unmapCloudIP(ctx, client, id, timeout)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Deleting Cloud IP %s", id)
	_, err = client.DestroyCloudIP(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func addUpdateableCloudipOptions(
	d *schema.ResourceData,
	opts *brightbox.CloudIPOptions,
) diag.Diagnostics {
	assignEnum(d, &opts.Mode, "mode")
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.ReverseDNS, "reverse_dns")
	assignPortTranslators(d, &opts.PortTranslators)
	return nil
}

func resourceBrightboxPortTranslationHash(
	v interface{},
) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["incoming"].(int)))
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["protocol"].(string))))
	buf.WriteString(fmt.Sprintf("%d-", m["outgoing"].(int)))

	return HashcodeString(buf.String())
}

func assignPortTranslators(d *schema.ResourceData, target *[]brightbox.PortTranslator) {
	if d.HasChange("port_translator") {
		*target = expandPortTranslators(d.Get("port_translator").(*schema.Set).List())
	}
}

func expandPortTranslators(configured []interface{}) []brightbox.PortTranslator {
	portTranslators := make([]brightbox.PortTranslator, len(configured))

	for i, portTranslationSource := range configured {
		data := portTranslationSource.(map[string]interface{})
		portTranslators[i].Protocol.UnmarshalText([]byte(data["protocol"].(string)))
		portTranslators[i].Incoming = data["incoming"].(uint16)
		portTranslators[i].Outgoing = data["outgoing"].(uint16)
	}
	return portTranslators
}
