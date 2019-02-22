package brightbox

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brightbox/gobrightbox"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBrightboxLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceBrightboxLoadBalancerCreate,
		Read:   resourceBrightboxLoadBalancerRead,
		Update: resourceBrightboxLoadBalancerUpdate,
		Delete: resourceBrightboxLoadBalancerDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: hash_string,
			},
			"certificate_private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: hash_string,
			},
			"sslv3": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"buffer_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"nodes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
			},
			"listener": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"in": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"out": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  50000,
						},
					},
				},
				Set: resourceBrightboxLbListenerHash,
			},
			"healthcheck": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"request": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"interval": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"threshold_up": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"threshold_down": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
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
	load_balancer *brightbox.LoadBalancer,
) error {
	d.Set("name", load_balancer.Name)
	d.Set("status", load_balancer.Status)
	d.Set("locked", load_balancer.Locked)
	d.Set("policy", load_balancer.Policy)
	d.Set("buffer_size", load_balancer.BufferSize)

	nodeIds := make([]string, 0, len(load_balancer.Nodes))
	for _, node := range load_balancer.Nodes {
		nodeIds = append(nodeIds, node.Id)
	}
	d.Set("nodes", nodeIds)

	cipIds := make([]string, 0, len(load_balancer.CloudIPs))
	for _, cip := range load_balancer.CloudIPs {
		cipIds = append(cipIds, cip.Id)
	}
	d.Set("cloud_ips", cipIds)

	listeners := make([]map[string]interface{}, 0, len(load_balancer.Listeners))
	for _, listener := range load_balancer.Listeners {
		l := map[string]interface{}{
			"protocol": listener.Protocol,
			"in":       listener.In,
			"out":      listener.Out,
			"timeout":  listener.Timeout,
		}

		listeners = append(listeners, l)
	}
	d.Set("listener", listeners)
	log.Printf("[DEBUG] Healthcheck details are %#v", load_balancer.Healthcheck)
	healthchecks := make([]map[string]interface{}, 0, 1)
	chk := map[string]interface{}{
		"type":           load_balancer.Healthcheck.Type,
		"port":           load_balancer.Healthcheck.Port,
		"request":        load_balancer.Healthcheck.Request,
		"interval":       load_balancer.Healthcheck.Interval,
		"timeout":        load_balancer.Healthcheck.Timeout,
		"threshold_up":   load_balancer.Healthcheck.ThresholdUp,
		"threshold_down": load_balancer.Healthcheck.ThresholdDown,
	}
	healthchecks = append(healthchecks, chk)
	d.Set("healthcheck", healthchecks)
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
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Load Balancer create called")
	load_balancer_opts := &brightbox.LoadBalancerOptions{}

	err := addUpdateableLoadBalancerOptions(d, load_balancer_opts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Load Balancer create configuration %#v", load_balancer_opts)
	output_load_balancer_options(load_balancer_opts)

	load_balancer, err := client.CreateLoadBalancer(load_balancer_opts)
	if err != nil {
		return fmt.Errorf("Error creating server: %s", err)
	}

	d.SetId(load_balancer.Id)

	log.Printf("[INFO] Waiting for Load Balancer (%s) to become available", d.Id())

	stateConf := resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    loadBalancerStateRefresh(client, load_balancer.Id),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	active_load_balancer, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	return setLoadBalancerAttributes(d, active_load_balancer.(*brightbox.LoadBalancer))
}

func resourceBrightboxLoadBalancerRead(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Load Balancer read called for %s", d.Id())
	load_balancer, err := client.LoadBalancer(d.Id())
	if err != nil {
		if strings.HasPrefix(err.Error(), "missing_resource:") {
			log.Printf("[WARN] Load Balancer not found, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Load Balancer details: %s", err)
	}

	return setLoadBalancerAttributes(d, load_balancer)
}

func resourceBrightboxLoadBalancerUpdate(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Load Balancer update called for %s", d.Id())
	load_balancer_opts := &brightbox.LoadBalancerOptions{
		Id: d.Id(),
	}

	err := addUpdateableLoadBalancerOptions(d, load_balancer_opts)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Load Balancer update configuration %#v", load_balancer_opts)
	output_load_balancer_options(load_balancer_opts)

	load_balancer, err := client.UpdateLoadBalancer(load_balancer_opts)
	if err != nil {
		return fmt.Errorf("Error updating load_balancer: %s", err)
	}

	return setLoadBalancerAttributes(d, load_balancer)
}

func resourceBrightboxLoadBalancerDelete(
	d *schema.ResourceData,
	meta interface{},
) error {
	client := meta.(*CompositeClient).ApiClient

	log.Printf("[DEBUG] Load Balancer delete called for %s", d.Id())
	err := client.DestroyLoadBalancer(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Load Balancer: %s", err)
	}
	stateConf := resource.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    loadBalancerStateRefresh(client, d.Id()),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
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
	assign_listeners(d, &opts.Listeners)
	assign_nodes(d, &opts.Nodes)
	return assign_healthcheck(d, &opts.Healthcheck)
}

func assign_healthcheck(d *schema.ResourceData, target **brightbox.LoadBalancerHealthcheck) error {
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

func assign_listeners(d *schema.ResourceData, target *[]brightbox.LoadBalancerListener) {
	if d.HasChange("listener") {
		*target = expandListeners(d.Get("listener").(*schema.Set).List())
	}
}

func assign_nodes(d *schema.ResourceData, target *[]brightbox.LoadBalancerNode) {
	if d.HasChange("nodes") {
		*target = expandNodes(d.Get("nodes").(*schema.Set).List())
	}
}

func expandListeners(configured []interface{}) []brightbox.LoadBalancerListener {
	listeners := make([]brightbox.LoadBalancerListener, len(configured))

	for i, listen_source := range configured {
		data := listen_source.(map[string]interface{})
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

func output_load_balancer_options(opts *brightbox.LoadBalancerOptions) {
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
