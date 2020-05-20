package brightbox

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	brightbox "github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceBrightboxLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxLoadBalancerCreate,
		Read:   resourceBrightboxLoadBalancerRead,
		Update: resourceBrightboxLoadBalancerUpdate,
		Delete: resourceBrightboxLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultTimeout),
			Delete: schema.DefaultTimeout(defaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Eitable user label",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"policy": {
				Description: "Method of load balancing. Supports `least-connections`, `round-robin` or `source-address`)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"certificate_pem": {
				Description: "A X509 SSL certificate in PEM format",
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   hash_string,
			},
			"certificate_private_key": {
				Description: "RSA private key used to sign the certificate in PEM format",
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc:   hash_string,
			},
			"sslv3": {
				Description: "Allow SSLv3 to be used",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"status": {
				Description: "Current state of the load balancer",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"locked": {
				Description: "Is true if resource has been set as locked and can not be deleted",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"buffer_size": {
				Description:  "Buffer size in bytes",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"nodes": {
				Description: "IDs of servers connected to this load balancer",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
				Set:         schema.HashString,
			},
			"listener": {
				Description: "Array of listeners",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Description: "The protocol to load balance (http/tcp)",
							Type:        schema.TypeString,
							Required:    true,
						},
						"in": {
							Description:  "The port this listener listens on",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(minPort, maxPort),
						},

						"out": {
							Description:  "The port on this server the listener should talk to",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(minPort, maxPort),
						},

						"timeout": {
							Description:  "Connection timeout in milliseconds",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      50000,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
				Set: resourceBrightboxLbListenerHash,
			},
			"healthcheck": {
				Description: "Healthcheck options",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "Protocol type to check (tcp/http)",
							Type:        schema.TypeString,
							Required:    true,
						},
						"port": {
							Description:  "Port on server to connect to for healthcheck",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(minPort, maxPort),
						},
						"request": {
							Description: "HTTP path to check if http type healthcheck",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"interval": {
							Description:  "How often to check in milliseconds",
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
						"threshold_up": {
							Description:  "How many checks have to pass before the load balancer considers the server active",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"threshold_down": {
							Description:  "How many checks have to fail before the load balancers considers a server inactive",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
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

	return hashcode.String(buf.String())
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

	nodeIds := make([]string, 0, len(loadBalancer.Nodes))
	for _, node := range loadBalancer.Nodes {
		nodeIds = append(nodeIds, node.Id)
	}
	if err := d.Set("nodes", nodeIds); err != nil {
		return fmt.Errorf("error setting nodes: %s", err)
	}

	listeners := make([]map[string]interface{}, len(loadBalancer.Listeners))
	for i, listener := range loadBalancer.Listeners {
		listeners[i] = map[string]interface{}{
			"protocol": listener.Protocol,
			"in":       listener.In,
			"out":      listener.Out,
			"timeout":  listener.Timeout,
		}
	}
	if err := d.Set("listener", listeners); err != nil {
		return fmt.Errorf("error setting listener: %s", err)
	}

	log.Printf("[DEBUG] Healthcheck details are %#v", loadBalancer.Healthcheck)
	healthchecks := make([]map[string]interface{}, 0, 1)
	chk := map[string]interface{}{
		"type":           loadBalancer.Healthcheck.Type,
		"port":           loadBalancer.Healthcheck.Port,
		"request":        loadBalancer.Healthcheck.Request,
		"interval":       loadBalancer.Healthcheck.Interval,
		"timeout":        loadBalancer.Healthcheck.Timeout,
		"threshold_up":   loadBalancer.Healthcheck.ThresholdUp,
		"threshold_down": loadBalancer.Healthcheck.ThresholdDown,
	}
	healthchecks = append(healthchecks, chk)
	if err := d.Set("healthcheck", healthchecks); err != nil {
		return fmt.Errorf("error setting healthcheck: %s", err)
	}

	log.Printf("[DEBUG] Certificate details are %#v", loadBalancer.Certificate)
	if loadBalancer.Certificate == nil {
		d.Set("sslv3", false)
	} else {
		d.Set("sslv3", loadBalancer.Certificate.SslV3)
	}
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
	if loadBalancer.Status == "deleted" {
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
	assign_string(d, &opts.Name, "name")
	assign_string(d, &opts.Policy, "policy")
	assign_string(d, &opts.CertificatePem, "certificate_pem")
	assign_string(d, &opts.CertificatePrivateKey, "certificate_private_key")
	assign_int(d, &opts.BufferSize, "buffer_size")
	assign_bool(d, &opts.SslV3, "sslv3")
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
	if opts.CertificatePem != nil {
		log.Printf("[DEBUG] Load Balancer CertificatePem %v", *opts.CertificatePem)
	}
	if opts.CertificatePrivateKey != nil {
		log.Printf("[DEBUG] Load Balancer CertificatePrivateKey %v", *opts.CertificatePrivateKey)
	}
	if opts.SslV3 != nil {
		log.Printf("[DEBUG] Load Balancer SslV3 %v", *opts.SslV3)
	}
}
