package brightbox

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	validSSLVersions           = []string{"TLSv1.0", "TLSv1.1", "TLSv1.2", "TLSv1.3", "SSLv3"}
	validListenerProtocols     = []string{"tcp", "http", "https", "http+ws", "https+wss"}
	validHealthcheckType       = []string{"tcp", "http"}
	validLoadBalancingPolicies = []string{"least-connections", "round-robin", "source-address"}
)

const (
	defaultListenerTimeout = 50000
)

func resourceBrightboxLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Brightbox Load Balancer resource",
		Create:      resourceBrightboxLoadBalancerCreate,
		Read:        resourceBrightboxLoadBalancerRead,
		Update:      resourceBrightboxLoadBalancerUpdate,
		Delete:      resourceBrightboxLoadBalancerDelete,
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
								validHealthcheckType,
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
								validListenerProtocols,
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
					validLoadBalancingPolicies,
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

			"ssl_minimum_version": {
				Description: "The minimum TLS/SSL version for the load balancer to accept. Supports `TLSv1.0`, TLSv1.1`, `TLSv1.2`, `TLSv1.3` and `SSLv3`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice(
					validSSLVersions,
					false,
				),
			},

			"status": {
				Description: "Current state of the load balancer",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
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
	buf.WriteString(fmt.Sprintf("%d-", m["out"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["timeout"].(int)))

	return HashcodeString(buf.String())
}

func mapFromListeners(
	listenerSet []brightbox.LoadBalancerListener,
) []map[string]interface{} {
	listeners := make([]map[string]interface{}, 0, len(listenerSet))
	for _, listener := range listenerSet {
		listeners = append(
			listeners,
			map[string]interface{}{
				"protocol": listener.Protocol,
				"in":       listener.In,
				"out":      listener.Out,
				"timeout":  listener.Timeout,
			},
		)
	}
	return listeners
}

func mapFromHealthcheck(
	healthcheck *brightbox.LoadBalancerHealthcheck,
) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":           healthcheck.Type,
			"port":           healthcheck.Port,
			"request":        healthcheck.Request,
			"interval":       healthcheck.Interval,
			"timeout":        healthcheck.Timeout,
			"threshold_up":   healthcheck.ThresholdUp,
			"threshold_down": healthcheck.ThresholdDown,
		},
	}

}

func setLoadBalancerAttributes(
	d *schema.ResourceData,
	loadBalancer *brightbox.LoadBalancer,
) error {
	d.Set("name", loadBalancer.Name)
	d.Set("status", loadBalancer.Status)
	d.Set("locked", loadBalancer.Locked)
	d.Set("policy", loadBalancer.Policy)
	d.Set("buffer_size", loadBalancer.BufferSize)
	d.Set("https_redirect", loadBalancer.HttpsRedirect)
	d.Set("ssl_minimum_version", loadBalancer.SslMinimumVersion)
	d.Set("sslv3", false)

	if err := d.Set("nodes", serverIDListFromNodes(loadBalancer.Nodes)); err != nil {
		return fmt.Errorf("error setting nodes: %s", err)
	}

	if err := d.Set("listener", mapFromListeners(loadBalancer.Listeners)); err != nil {
		return fmt.Errorf("error setting listener: %s", err)
	}

	log.Printf("[DEBUG] Healthcheck details are %#v", loadBalancer.Healthcheck)
	if err := d.Set("healthcheck", mapFromHealthcheck(&loadBalancer.Healthcheck)); err != nil {
		return fmt.Errorf("error setting healthcheck: %s", err)
	}

	log.Printf("[DEBUG] Certificate details are %#v", loadBalancer.Certificate)

	return nil
}

func loadBalancerStateRefresh(client *brightbox.Client, loadBalancerID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		loadBalancer, err := client.LoadBalancer(loadBalancerID)
		if err != nil {
			log.Printf("Error on Load Balancer State Refresh: %s", err)
			return nil, "", err
		}
		return loadBalancer, loadBalancer.Status, nil
	}
}

func resourceBrightboxLoadBalancerCreate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Load Balancer create called")
	loadBalancerOpts := &brightbox.LoadBalancerOptions{}

	err := addUpdateableLoadBalancerOptions(d, loadBalancerOpts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Load Balancer create configuration %#v", loadBalancerOpts)
	outputLoadBalancerOptions(loadBalancerOpts)

	loadBalancer, err := client.CreateLoadBalancer(loadBalancerOpts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(loadBalancer.Id)

	locked := d.Get("locked").(bool)
	log.Printf("[INFO] Setting lock state to %v", locked)
	if err := setLockState(client, locked, loadBalancer); err != nil {
		return err
	}

	log.Printf("[INFO] Waiting for Load Balancer (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    loadBalancerStateRefresh(client, loadBalancer.Id),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	activeLoadBalancer, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	return setLoadBalancerAttributes(d, activeLoadBalancer.(*brightbox.LoadBalancer))
}

func resourceBrightboxLoadBalancerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Load Balancer read called for %s", d.Id())
	loadBalancer, err := client.LoadBalancer(d.Id())
	if err != nil {
		return fmt.Errorf("Error retrieving Load Balancer details: %s", err)
	}
	if unreadable[loadBalancer.Status] {
		log.Printf("[WARN] Load Balancer not found, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	return setLoadBalancerAttributes(d, loadBalancer)
}

func resourceBrightboxLoadBalancerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Load Balancer update called for %s", d.Id())
	loadBalancerOpts := &brightbox.LoadBalancerOptions{
		Id: d.Id(),
	}

	err := addUpdateableLoadBalancerOptions(d, loadBalancerOpts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Load Balancer update configuration %#v", loadBalancerOpts)
	outputLoadBalancerOptions(loadBalancerOpts)

	loadBalancer, err := client.UpdateLoadBalancer(loadBalancerOpts)
	if err != nil {
		return fmt.Errorf("Error updating loadBalancer: %s", err)
	}

	if d.HasChange("locked") {
		locked := d.Get("locked").(bool)
		log.Printf("[INFO] Setting lock state to %v", locked)
		if err := setLockState(client, locked, loadBalancer); err != nil {
			return err
		}
		return resourceBrightboxLoadBalancerRead(d, meta)
	}
	return setLoadBalancerAttributes(d, loadBalancer)
}

func resourceBrightboxLoadBalancerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).APIClient

	log.Printf("[DEBUG] Load Balancer delete called for %s", d.Id())
	err := client.DestroyLoadBalancer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Load Balancer: %s", err)
	}
	stateConf := resource.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    loadBalancerStateRefresh(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      checkDelay,
		MinTimeout: minimumRefreshWait,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	return nil
}

func addUpdateableLoadBalancerOptions(
	d *schema.ResourceData,
	opts *brightbox.LoadBalancerOptions,
) error {
	assignString(d, &opts.Name, "name")
	assignString(d, &opts.Policy, "policy")
	assignString(d, &opts.CertificatePem, "certificate_pem")
	assignString(d, &opts.CertificatePrivateKey, "certificate_private_key")
	assignString(d, &opts.SslMinimumVersion, "ssl_minimum_version")
	assignInt(d, &opts.BufferSize, "buffer_size")
	assignBool(d, &opts.HttpsRedirect, "https_redirect")
	assignStringSet(d, &opts.Domains, "domains")
	assignListeners(d, &opts.Listeners)
	assignNodes(d, &opts.Nodes)
	return assignHealthCheck(d, &opts.Healthcheck)
}

func assignHealthCheck(d *schema.ResourceData, target **brightbox.LoadBalancerHealthcheck) error {
	if d.HasChange("healthcheck") {
		hc := d.Get("healthcheck").([]interface{})
		check := hc[0].(map[string]interface{})
		temp := brightbox.LoadBalancerHealthcheck{
			Type: check["type"].(string),
			Port: check["port"].(int),
		}
		if attr, ok := check["request"]; ok {
			temp.Request = attr.(string)
		}
		if attr, ok := check["interval"]; ok {
			temp.Interval = attr.(int)
		}
		if attr, ok := check["timeout"]; ok {
			temp.Timeout = attr.(int)
		}
		if attr, ok := check["threshold_up"]; ok {
			temp.ThresholdUp = attr.(int)
		}
		if attr, ok := check["threshold_down"]; ok {
			temp.ThresholdDown = attr.(int)
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
		listeners[i].Protocol = data["protocol"].(string)
		listeners[i].In = data["in"].(int)
		listeners[i].Out = data["out"].(int)
		if attr, ok := data["timeout"]; ok {
			listeners[i].Timeout = attr.(int)
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
	if opts.Policy != nil {
		log.Printf("[DEBUG] Load Balancer Policy %v", *opts.Policy)
	}
	if opts.BufferSize != nil {
		log.Printf("[DEBUG] Load Balancer BufferSize %v", *opts.BufferSize)
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
