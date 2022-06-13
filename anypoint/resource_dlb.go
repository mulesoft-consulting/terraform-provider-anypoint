package anypoint

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iancoleman/strcase"
	"github.com/mulesoft-consulting/anypoint-client-go/dlb"
)

func resourceDLB() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDLBCreate,
		ReadContext:   resourceDLBRead,
		UpdateContext: resourceDLBUpdate,
		DeleteContext: resourceDLBDelete,
		Description: `
		Creates a ` + "`" + `dedicated load balancer` + "`" + ` instance in your ` + "`" + `vpc` + "`" + `.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the DLB.",
			},
			"domain": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The DNS domain for the Load Balancer",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "" {
						return true
					} else {
						return old == new
					}
				},
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "stopped",
				Description: "The desired state, possible values: 'started', 'stopped' or 'restarted'",
				// Suppress the diff shown if the state name are equal when both compared in lower case.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return compareDLBStates(old, new)
				},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"started", "stopped", "restarted"}, true)),
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_config": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"ip_addresses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of static IP addresses for the Load Balancer",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_whitelist": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "CIDR blocks to allow connections from",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_allowlist": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "CIDR blocks to allow connections from",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"http_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "redirect",
				Description:      "Specifies whether the Load Balancer listens for HTTP requests on port 80. If set to redirect, all HTTP requests will be redirected to HTTPS. possible values: 'on', 'off' or 'redirect'",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"on", "off", "redirect"}, true)),
			},
			"default_ssl_endpoint": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The default certificate that will be served for requests not using SNI, or requesting a non-existing certificate",
			},
			"ssl_endpoints": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"public_key": {
							Type:      schema.TypeString,
							Sensitive: true,
							Required:  true,
						},
						"private_key": {
							Type:      schema.TypeString,
							Sensitive: true,
							Required:  true,
						},
						"private_key_digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key_label": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"public_key_digest": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key_cn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_key_label": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"verify_client_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "off",
							Description:      "Whether to enable client verification or not, possible values: 'off' or 'on'",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"off", "on"}, true)),
						},
						"mappings": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"app_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"app_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"upstream_protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"static_ips_disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"workers": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_cipher_suite": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"keep_url_encoding": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tlsv1": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"upstream_tlsv12": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"proxy_read_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ip_addresses_info": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of IP addresses information for the Load Balancer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"static_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"double_static_ips": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True if DLB will use double static IPs when restarting",
			},
			"enable_streaming": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Setting this to true will disable request buffering at the DLB, thereby enabling streaming",
			},
			"forward_client_certificate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Setting this to true will forward any incoming client certificates to upstream application",
			},
		},
	}
}

func resourceDLBCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)
	body := newDLBPostBody(d)

	//request user creation
	res, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersPost(authctx, orgid, vpcid).DlbPostBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create DLB of org " + orgid + " and vpc " + vpcid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())

	resourceDLBRead(ctx, d, m)

	return diags
}

func resourceDLBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Id()
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)

	//request roles
	res, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdGet(authctx, orgid, vpcid, dlbid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get dlb " + dlbid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	dlb := flattenDLBData(&res)
	//save in data source schema
	if err := setDLBAttributesToResourceData(d, dlb); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set dlb " + dlbid,
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceDLBUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Id()
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)

	if d.HasChanges(getDLBPatchWatchAttributes()...) {
		body := newDLBPatchBody(d)
		//request user creation
		_, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdPatch(authctx, orgid, vpcid, dlbid).RequestBody(body).Execute()
		if err != nil {
			var details string
			if httpr != nil {
				b, _ := ioutil.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to patch dlb " + dlbid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceDLBRead(ctx, d, m)
}

func resourceDLBDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Id()
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)

	httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdDelete(authctx, orgid, vpcid, dlbid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete dlb " + dlbid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func newDLBPostBody(d *schema.ResourceData) *dlb.DlbPostBody {
	body := dlb.NewDlbPostBody()
	if name := d.Get("name"); name != nil {
		body.SetName(name.(string))
	}
	if state := d.Get("state"); state != nil {
		body.SetState(state.(string))
	}
	if domain := d.Get("domain"); domain != nil {
		body.SetDomain(domain.(string))
	}
	if ip_whitelist := d.Get("ip_whitelist"); ip_whitelist != nil {
		body.SetIpWhitelist(ListInterface2ListStrings(ip_whitelist.([]interface{})))
	}
	if ip_allowlist := d.Get("ip_allowlist"); ip_allowlist != nil {
		body.SetIpAllowlist(ListInterface2ListStrings(ip_allowlist.([]interface{})))
	}
	if http_mode := d.Get("http_mode"); http_mode != nil {
		body.SetHttpMode(http_mode.(string))
	}
	if keep_url_encoding := d.Get("keep_url_encoding"); keep_url_encoding != nil {
		body.SetKeepUrlEncoding(keep_url_encoding.(bool))
	}
	if upstream_tlsv12 := d.Get("upstream_tlsv12"); upstream_tlsv12 != nil {
		body.SetUpstreamTlsv12(upstream_tlsv12.(bool))
	}
	if tlsv1 := d.Get("tlsv1"); tlsv1 != nil {
		body.SetTlsv1(tlsv1.(bool))
	}
	if double_static_ips := d.Get("double_static_ips"); double_static_ips != nil {
		body.SetDoubleStaticIps(double_static_ips.(bool))
	}
	if enable_streaming := d.Get("enable_streaming"); enable_streaming != nil {
		body.SetEnableStreaming(enable_streaming.(bool))
	}
	if forward_client_certificate := d.Get("forward_client_certificate"); forward_client_certificate != nil {
		body.SetForwardClientCertificate(forward_client_certificate.(bool))
	}
	if default_ssl_endpoint := d.Get("default_ssl_endpoint"); default_ssl_endpoint != nil {
		body.SetDefaultSslEndpoint(int32(default_ssl_endpoint.(int)))
	}
	if ssl_endpoints := d.Get("ssl_endpoints"); ssl_endpoints != nil {
		ssl_endpoints_set := ssl_endpoints.(*schema.Set)
		ssl_endpoints_list := ssl_endpoints_set.List()
		ssl_endpoints_body := make([]dlb.DlbPostBodySslEndpoints, len(ssl_endpoints_list))
		for i, endpoint := range ssl_endpoints_list {
			endpoint_converted := endpoint.(map[string]interface{})
			endpoint_item := dlb.NewDlbPostBodySslEndpoints()
			if val, ok := endpoint_converted["public_key"]; ok {
				endpoint_item.SetPublicKey(val.(string))
			}
			if val, ok := endpoint_converted["private_key"]; ok {
				endpoint_item.SetPrivateKey(val.(string))
			}
			if val, ok := endpoint_converted["public_key_label"]; ok {
				endpoint_item.SetPublicKeyLabel(val.(string))
			}
			if val, ok := endpoint_converted["private_key_label"]; ok {
				endpoint_item.SetPrivateKeyLabel(val.(string))
			}
			if val, ok := endpoint_converted["verify_client_mode"]; ok {
				endpoint_item.SetVerifyClientMode(val.(string))
			}
			if val, ok := endpoint_converted["mappings"]; ok {
				mappings := val.([]interface{})
				mappings_body := make([]dlb.DlbPostBodyMappings, len(mappings))
				for j, mapping := range mappings {
					mapping_converted := mapping.(map[string]interface{})
					mapping_item := dlb.NewDlbPostBodyMappings()
					if val, ok := mapping_converted["input_uri"]; ok {
						mapping_item.SetInputUri(val.(string))
					}
					if val, ok := mapping_converted["app_name"]; ok {
						mapping_item.SetAppName(val.(string))
					}
					if val, ok := mapping_converted["app_uri"]; ok {
						mapping_item.SetAppUri(val.(string))
					}
					mappings_body[j] = *mapping_item
				}
				endpoint_item.SetMappings(mappings_body)
			}
			ssl_endpoints_body[i] = *endpoint_item
		}
		body.SetSslEndpoints(ssl_endpoints_body)
	}
	return body
}

func newDLBPatchBody(d *schema.ResourceData) []map[string]interface{} {
	attributes := getDLBPatchWatchAttributes()
	body := make([]map[string]interface{}, len(attributes))
	op_replace := "replace"
	for i, attr := range attributes {
		camlAttr := strcase.ToLowerCamel(attr)
		item := make(map[string]interface{})
		if attr == "ssl_endpoints" {
			ssl_endpoints_set := d.Get(attr).(*schema.Set)
			ssl_endpoints_list := ssl_endpoints_set.List()
			ssl_endpoints_extract := make([]map[string]interface{}, len(ssl_endpoints_list))
			for j, val := range ssl_endpoints_list {
				endpoint := val.(map[string]interface{})
				e := make(map[string]interface{})
				public_key_field := "public_key"
				public_key_label_field := "public_key_label"
				private_key_field := "private_key"
				private_key_label_field := "private_key_label"
				verify_client_mode_field := "verify_client_mode"
				mappings_field := "mappings"
				e[strcase.ToLowerCamel(public_key_field)] = endpoint[public_key_field].(string)
				e[strcase.ToLowerCamel(public_key_label_field)] = endpoint[public_key_label_field].(string)
				e[strcase.ToLowerCamel(private_key_field)] = endpoint[private_key_field].(string)
				e[strcase.ToLowerCamel(private_key_label_field)] = endpoint[private_key_label_field].(string)
				e[strcase.ToLowerCamel(verify_client_mode_field)] = endpoint[verify_client_mode_field].(string)
				mappings := endpoint[mappings_field].([]interface{})
				mappings_extract := make([]map[string]interface{}, len(mappings))
				for k, mappings_val := range mappings {
					mapping := mappings_val.(map[string]interface{})
					m := make(map[string]interface{})
					input_uri_field := "input_uri"
					app_name_field := "app_name"
					app_uri_field := "app_uri"
					m[strcase.ToLowerCamel(input_uri_field)] = mapping[input_uri_field].(string)
					m[strcase.ToLowerCamel(app_name_field)] = mapping[app_name_field].(string)
					m[strcase.ToLowerCamel(app_uri_field)] = mapping[app_uri_field].(string)
					mappings_extract[k] = m
				}
				e[strcase.ToLowerCamel(mappings_field)] = mappings_extract
				ssl_endpoints_extract[j] = e
			}
			item["op"] = op_replace
			item["path"] = "/" + camlAttr
			item["value"] = ssl_endpoints_extract
		} else if StringInSlice([]string{"ip_whitelist", "ip_allowlist"}, attr, false) { // update of
			item["op"] = op_replace
			item["path"] = "/" + camlAttr
			item["value"] = ListInterface2ListStrings(d.Get(attr).([]interface{}))
		} else if StringInSlice([]string{
			"tlsv1", "upstream_tlsv12", "keep_url_encoding",
			"double_static_ips", "enable_streaming", "forward_client_certificate",
		}, attr, false) {
			item["op"] = op_replace
			item["path"] = "/" + camlAttr
			item["value"] = d.Get(attr).(bool)
		} else if StringInSlice([]string{"default_ssl_endpoint"}, attr, false) {
			item["op"] = op_replace
			item["path"] = "/" + camlAttr
			item["value"] = d.Get(attr).(int)
		} else {
			item["op"] = op_replace
			item["path"] = "/" + camlAttr
			item["value"] = d.Get(attr).(string)
		}
		body[i] = item
	}
	return body
}

func getDLBPatchWatchAttributes() []string {
	attributes := [...]string{
		"state", "ip_whitelist", "ip_allowlist", "http_mode",
		"default_ssl_endpoint", "ssl_endpoints", "tlsv1", "upstream_tlsv12",
		"keep_url_encoding", "double_static_ips", "enable_streaming",
		"forward_client_certificate",
	}
	return attributes[:]
}

// Compares DLB states
func compareDLBStates(old, new string) bool {
	old_lowercase := strings.ToLower(old)
	new_lowercase := strings.ToLower(new)
	if strings.EqualFold(old, new) {
		return true
	} else if new_lowercase == "started" {
		if old_lowercase == "starting" {
			return true
		}
	} else if new_lowercase == "stopped" {
		if old_lowercase == "stopping" {
			return true
		}
	} else if new_lowercase == "restarted" {
		if old_lowercase == "restarting" {
			return true
		}
		if old_lowercase == "updating" {
			return true
		}
	}
	return false
}

/*
 * Returns authentication context (includes authorization header)
 */
func getDLBAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, dlb.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, dlb.ContextServerIndex, pco.server_index)
}
