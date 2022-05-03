package brightbox

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/brightbox/gobrightbox/v2/enums/balancingpolicy"
	"github.com/brightbox/gobrightbox/v2/enums/healthchecktype"
	"github.com/brightbox/gobrightbox/v2/enums/listenerprotocol"
	"github.com/brightbox/gobrightbox/v2/enums/loadbalancerstatus"
	"github.com/brightbox/gobrightbox/v2/enums/proxyprotocol"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	validSSLVersions = []string{"TLSv1.0", "TLSv1.1", "TLSv1.2", "TLSv1.3", "SSLv3"}
)

const (
	defaultListenerTimeout = 50000
)

func resourceBrightboxLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Brightbox Load Balancer resource",
		CreateContext: resourceBrightboxLoadBalancerCreateAndWait,
		ReadContext:   resourceBrightboxLoadBalancerRead,
		UpdateContext: resourceBrightboxLoadBalancerUpdate,
		DeleteContext: resourceBrightboxLoadBalancerDeleteAndWait,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{

			"buffer_size": {
				Description:  "Buffer size in bytes",
				Type:         schema.TypeInt,
				Deprecated:   "No longer supported. Buffer size is automatically calculated",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"certificate_pem": {
				Description: "A X509 SSL certificate in PEM format",
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   hashString,
			},

			"certificate_private_key": {
				Description: "RSA private key used to sign the certificate in PEM format",
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   hashString,
			},

			"domains": {
				Description: "Array of domain names to attempt to register with ACME",
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(dnsNameRegexp, "must be a valid DNS name"),
				},
				ConflictsWith: []string{"certificate_pem", "certificate_private_key"},
			},

			"healthcheck": {
				Description: "Healthcheck options",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						"interval": {
							Description:  "How often to check in milliseconds",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"port": {
							Description:  "Port on server to connect to for healthcheck",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"request": {
							Description:  "HTTP path to check if http type healthcheck",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringIsNotWhiteSpace,
						},

						"threshold_down": {
							Description:  "How many checks have to fail before the load balancers considers a server inactive",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"threshold_up": {
							Description:  "How many checks have to pass before the load balancer considers the server active",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"timeout": {
							Description:  "How long to wait for a response before marking the check as a fail",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"type": {
							Description: "Protocol type to check (tcp/http)",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								healthchecktype.ValidStrings,
								false,
							),
						},
					},
				},
			},

			"https_redirect": {
				Description: "Redirect any requests on port 80 automatically to port 443",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"listener": {
				Description: "Array of listeners",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"in": {
							Description:  "The port this listener listens on",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"out": {
							Description:  "The port on this server the listener should talk to",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"protocol": {
							Description: "The protocol to load balance (http/tcp)",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: validation.StringInSlice(
								listenerprotocol.ValidStrings,
								false,
							),
						},

						"proxy_protocol": {
							Description: "The version of the Proxy Protocol supported by the backend servers",
							Type:        schema.TypeString,
							Optional:    true,
							ValidateFunc: validation.StringInSlice(
								proxyprotocol.ValidStrings,
								false,
							),
						},

						"timeout": {
							Description:  "Connection timeout in milliseconds",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      defaultListenerTimeout,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
				Set: resourceBrightboxLbListenerHash,
			},

			"locked": {
				Description: "Is true if resource has been set as locked and can not be deleted",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"name": {
				Description: "Editable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"nodes": {
				Description: "IDs of servers connected to this load balancer",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(serverRegexp, "must be a valid server ID"),
				},
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
			},

			"policy": {
				Description: "Method of load balancing. Supports `least-connections`, `round-robin` or `source-address`)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				// Default:     "least-connections",
				ValidateFunc: validation.StringInSlice(
					balancingpolicy.ValidStrings,
					false,
				),
			},

			"ssl_minimum_version": {
				Description: "The minimum TLS/SSL version for the load balancer to accept. Supports `TLSv1.0`, `TLSv1.1`, `TLSv1.2`, `TLSv1.3` and `SSLv3`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validSSLVersions,
					false,
				),
			},

			"sslv3": {
				Description: "Allow SSLv3 to be used",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Deprecated:  "No longer supported. Will always return false",
			},

			"status": {
				Description: "Current state of the load balancer",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

var (
	resourceBrightboxSetLoadBalancerLockState = resourceBrightboxSetLockState(
		(*brightbox.Client).LockLoadBalancer,
		(*brightbox.Client).UnlockLoadBalancer,
		setLoadBalancerAttributes,
	)

	resourceBrightboxLoadBalancerRead = resourceBrightboxReadStatus(
		(*brightbox.Client).LoadBalancer,
		"Load Balancer",
		setLoadBalancerAttributes,
		loadBalancerUnavailable,
	)

	resourceBrightboxLoadBalancerUpdate = resourceBrightboxUpdateWithLock(
		(*brightbox.Client).UpdateLoadBalancer,
		"Load Balancer",
		loadBalancerFromID,
		addUpdateableLoadBalancerOptions,
		setLoadBalancerAttributes,
		resourceBrightboxSetLoadBalancerLockState,
	)

	resourceBrightboxLoadBalancerDeleteAndWait = resourceBrightboxDeleteAndWait(
		(*brightbox.Client).DestroyLoadBalancer,
		"Load Balancer",
		[]string{
			loadbalancerstatus.Deleting.String(),
			loadbalancerstatus.Active.String(),
		},
		[]string{
			loadbalancerstatus.Deleted.String(),
		},
		loadBalancerStateRefresh,
	)
)

func addUpdateableLoadBalancerOptions(
	d *schema.ResourceData,
	opts *brightbox.LoadBalancerOptions,
) diag.Diagnostics {
	assignString(d, &opts.Name, "name")
	assignEnum(d, &opts.Policy, "policy")
	assignString(d, &opts.CertificatePem, "certificate_pem")
	assignString(d, &opts.CertificatePrivateKey, "certificate_private_key")
	assignString(d, &opts.SslMinimumVersion, "ssl_minimum_version")
	assignBool(d, &opts.HTTPSRedirect, "https_redirect")
	if d.HasChange("domains") {
		temp := sliceFromStringSet(d, "domains")
		opts.Domains = &temp
	}
	assignListeners(d, &opts.Listeners)
	assignNodes(d, &opts.Nodes)
	return assignHealthCheck(d, &opts.Healthcheck)
}

func setLoadBalancerAttributes(
	d *schema.ResourceData,
	loadBalancer *brightbox.LoadBalancer,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	d.SetId(loadBalancer.ID)
	err = d.Set("name", loadBalancer.Name)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("status", loadBalancer.Status.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("locked", loadBalancer.Locked)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("policy", loadBalancer.Policy.String())
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("buffer_size", loadBalancer.BufferSize)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("https_redirect", loadBalancer.HTTPSRedirect)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("ssl_minimum_version", loadBalancer.SslMinimumVersion)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("sslv3", false)
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("domains", stringSliceFromAcme(loadBalancer.Acme))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("nodes", serverIDListFromNodes(loadBalancer.Nodes))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	err = d.Set("listener", mapFromListeners(loadBalancer.Listeners))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	log.Printf("[DEBUG] Healthcheck details are %+v", loadBalancer.Healthcheck)
	err = d.Set("healthcheck", mapFromHealthcheck(&loadBalancer.Healthcheck))
	if err != nil {
		diags = append(diags, diag.Errorf("unexpected: %s", err)...)
	}
	log.Printf("[DEBUG] Certificate details are %+v", loadBalancer.Certificate)
	return diags
}

func loadBalancerFromID(id string) *brightbox.LoadBalancerOptions {
	return &brightbox.LoadBalancerOptions{
		ID: id,
	}
}

func loadBalancerUnavailable(obj *brightbox.LoadBalancer) bool {
	return obj.Status == loadbalancerstatus.Deleted ||
		obj.Status == loadbalancerstatus.Failed
}

func loadBalancerStateRefresh(client *brightbox.Client, ctx context.Context, loadBalancerID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		loadBalancer, err := client.LoadBalancer(ctx, loadBalancerID)
		if err != nil {
			log.Printf("Error on Load Balancer State Refresh: %s", err)
			return nil, "", err
		}
		return loadBalancer, loadBalancer.Status.String(), nil
	}
}

func resourceBrightboxLbListenerHash(
	v interface{},
) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["in"].(int)))
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["protocol"].(string))))
	buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["proxy_protocol"].(string))))
	buf.WriteString(fmt.Sprintf("%d-", m["out"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["timeout"].(int)))

	return HashcodeString(buf.String())
}

func stringSliceFromAcme(
	acme *brightbox.LoadBalancerAcme,
) []string {
	if acme == nil {
		return nil
	}
	result := make([]string, len(acme.Domains))
	for i, domain := range acme.Domains {
		result[i] = domain.Identifier
	}
	return result
}

func mapFromListeners(
	listenerSet []brightbox.LoadBalancerListener,
) []map[string]interface{} {
	listeners := make([]map[string]interface{}, len(listenerSet))
	for i, listener := range listenerSet {
		listeners[i] = map[string]interface{}{
			"protocol":       listener.Protocol.String(),
			"in":             listener.In,
			"out":            listener.Out,
			"timeout":        listener.Timeout,
			"proxy_protocol": listener.ProxyProtocol.String(),
		}
	}
	return listeners
}

func mapFromHealthcheck(
	healthcheck *brightbox.LoadBalancerHealthcheck,
) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":           healthcheck.Type.String(),
			"port":           healthcheck.Port,
			"request":        healthcheck.Request,
			"interval":       healthcheck.Interval,
			"timeout":        healthcheck.Timeout,
			"threshold_up":   healthcheck.ThresholdUp,
			"threshold_down": healthcheck.ThresholdDown,
		},
	}

}

func resourceBrightboxLoadBalancerCreateAndWait(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[INFO]] Creating Load Balancer")
	var loadBalancerOpts brightbox.LoadBalancerOptions

	errs := addUpdateableLoadBalancerOptions(d, &loadBalancerOpts)
	if errs.HasError() {
		return errs
	}

	log.Printf("[DEBUG] Load Balancer create configuration %+v", loadBalancerOpts)
	outputLoadBalancerOptions(&loadBalancerOpts)

	loadBalancer, err := client.CreateLoadBalancer(ctx, loadBalancerOpts)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(loadBalancer.ID)

	log.Printf("[INFO] Waiting for Load Balancer (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending: []string{
			loadbalancerstatus.Creating.String(),
		},
		Target: []string{
			loadbalancerstatus.Active.String(),
		},
		Refresh:    loadBalancerStateRefresh(client, ctx, loadBalancer.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBrightboxSetLoadBalancerLockState(ctx, d, meta)
}

func assignHealthCheck(d *schema.ResourceData, target **brightbox.LoadBalancerHealthcheck) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.HasChange("healthcheck") {
		hc := d.Get("healthcheck").([]interface{})
		check := hc[0].(map[string]interface{})
		temp := brightbox.LoadBalancerHealthcheck{
			Port: uint16(check["port"].(int)),
		}
		if hctype, err := healthchecktype.ParseEnum(check["type"].(string)); err == nil {
			temp.Type = hctype
		} else {
			diags = append(diags, diag.FromErr(err)...)
		}
		if attr, ok := check["request"]; ok {
			temp.Request = attr.(string)
		}
		if attr, ok := check["interval"]; ok {
			temp.Interval = uint(attr.(int))
		}
		if attr, ok := check["timeout"]; ok {
			temp.Timeout = uint(attr.(int))
		}
		if attr, ok := check["threshold_up"]; ok {
			temp.ThresholdUp = uint(attr.(int))
		}
		if attr, ok := check["threshold_down"]; ok {
			temp.ThresholdDown = uint(attr.(int))
		}
		*target = &temp
	}
	return nil
}

func assignListeners(d *schema.ResourceData, target *[]brightbox.LoadBalancerListener) {
	if d.HasChange("listener") {
		*target = expandListeners(d.Get("listener").(*schema.Set).List())
	}
}

func assignNodes(d *schema.ResourceData, target *[]brightbox.LoadBalancerNode) {
	if d.HasChange("nodes") {
		*target = expandNodes(d.Get("nodes").(*schema.Set).List())
	}
}

func expandListeners(configured []interface{}) []brightbox.LoadBalancerListener {
	listeners := make([]brightbox.LoadBalancerListener, len(configured))

	for i, listenSource := range configured {
		data := listenSource.(map[string]interface{})
		protocolTarget := &listeners[i].Protocol
		protocolTarget.UnmarshalText([]byte(data["protocol"].(string)))
		listeners[i].In = uint16(data["in"].(int))
		listeners[i].Out = uint16(data["out"].(int))
		if attr, ok := data["timeout"]; ok {
			listeners[i].Timeout = uint(attr.(int))
		}
		if attr, ok := data["proxy_protocol"]; ok {
			proxyTarget := &listeners[i].ProxyProtocol
			proxyTarget.UnmarshalText([]byte(attr.(string)))
		}
	}
	return listeners
}

func expandNodes(configured []interface{}) []brightbox.LoadBalancerNode {
	nodes := make([]brightbox.LoadBalancerNode, len(configured))

	for i, data := range configured {
		nodes[i].Node = data.(string)
	}
	return nodes
}

func outputLoadBalancerOptions(opts *brightbox.LoadBalancerOptions) {
	if opts.Name != nil {
		log.Printf("[DEBUG] Load Balancer Name %v", *opts.Name)
	}
	if opts.Nodes != nil {
		log.Printf("[DEBUG] Load Balancer Nodes %#v", opts.Nodes)
	}
	if opts.Policy != 0 {
		log.Printf("[DEBUG] Load Balancer Policy %v", opts.Policy.String())
	}
	if opts.Listeners != nil {
		log.Printf("[DEBUG] Load Balancer Listeners %#v", opts.Listeners)
	}
	if opts.Healthcheck != nil {
		log.Printf("[DEBUG] Load Balancer Healthcheck %#v", *opts.Healthcheck)
	}
	if opts.Domains != nil {
		log.Printf("[DEBUG] Load Balancer Domains %#v", opts.Domains)
	}
	if opts.CertificatePem != nil {
		log.Printf("[DEBUG] Load Balancer CertificatePem %v", *opts.CertificatePem)
	}
	if opts.CertificatePrivateKey != nil {
		log.Printf("[DEBUG] Load Balancer CertificatePrivateKey %v", *opts.CertificatePrivateKey)
	}
}
