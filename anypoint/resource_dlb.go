package anypoint

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iancoleman/strcase"
	"github.com/mulesoft-anypoint/anypoint-client-go/dlb"
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of this dlb generated by the anypoint platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The organization id where the dlb is defined.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The vpc id",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the dlb.",
			},
			"domain": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The domain name of this dlb",
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
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"started", "stopped", "restarted", "starting", "updating", "stopping"}, true)),
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
				Description: "List of static IP addresses for this dlb",
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
				ConflictsWith: []string{"ip_allowlist"},
				//checks wether the whitelist has changed
				//uses custom function in case ip_allowlist is used instead, then this field is ignored
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalDLBAllowList(d.GetChange("ip_whitelist"))
				},
			},
			"ip_allowlist": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "CIDR blocks to allow connections from",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"ip_whitelist"},
				//checks wether the allowlist has changed
				//uses custom function in case ip_whitelist is used instead, then this field is ignored
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalDLBAllowList(d.GetChange("ip_allowlist"))
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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return equalDLBSSLEndpoints(d.GetChange("ssl_endpoints"))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_key": {
							Type:        schema.TypeString,
							Sensitive:   true,
							Required:    true,
							Description: "The private key of the given endpoint",
						},
						"private_key_digest": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private key checksum generated by the anypoint platform.",
						},
						"private_key_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The label of the private key.",
						},
						"public_key": {
							Type:        schema.TypeString,
							Sensitive:   true,
							Required:    true,
							Description: "The public key of the given endpoint.",
						},
						"public_key_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The label of the public key.",
						},
						"public_key_digest": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public key checksum generated by the anypoint platform.",
						},
						"public_key_cn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The common name of the public key.",
						},
						"client_cert": {
							Type:        schema.TypeString,
							Sensitive:   true,
							Optional:    true,
							Description: "The client certificat of the given endpoint",
						},
						"client_cert_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The label of the client certificat.",
						},
						"client_cert_digest": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The client certificate checksum generated by the anypoint platform.",
						},
						"client_cert_cn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The common name of the client's certificate.",
						},
						"revocation_list": {
							Type:        schema.TypeString,
							Sensitive:   true,
							Optional:    true,
							Description: "The revocation list for the given endpoint",
						},
						"revocation_list_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The label of the revocation list.",
						},
						"revocation_list_digest": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The CRL checksum generated by the anypoint platform.",
						},
						"verify_client_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "off",
							Description:      "Whether to enable client verification or not, possible values: 'off' or 'on' or 'optional'",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"off", "on", "optional"}, true)),
						},
						"mappings": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "List of dlb mappings.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_uri": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The URI that the client requests: for example, `/{app}/`. The input URI is appended to the host header of the load balancer.",
									},
									"app_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the CloudHub application that processes the request: for example, {app}-example",
									},
									"app_uri": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The URI string to pass to the app: for example, `/`. The output path cannot contain patterns.",
									},
									"upstream_protocol": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "http",
										Description: `
										The protocol on which the application listens:
											* http (port 8091)
											* https (port 8092)
											* ws WebSockets (port 8091)
											* wss WebSockets over SSL/TLS (8092)
											By default, the load balancer listens to external requests on HTTPS and communicates internally with your worker over HTTP. If you configured your Mule application within the VPC to listen on HTTPS, set Protocol to https when you create the mapping rule list.
										`,
										ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"http", "https", "ws", "wss"}, true)),
									},
								},
							},
						},
					},
				},
			},
			"static_ips_disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to disable static ips for this dlb.",
			},
			"workers": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          2,
				Description:      "The number of workers for this dlb.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntInSlice([]int{2, 4, 6, 8})),
			},
			"default_cipher_suite": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The default cipher suite used by this dlb.",
			},
			"keep_url_encoding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to keep url encoding for this dlb.",
			},
			"tlsv1": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to activate TLS v1 for this dlb.",
			},
			"upstream_tlsv12": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to activate TLS v1.2 for this dlb upstream.",
			},
			"proxy_read_timeout": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          300,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
				Description:      "The proxy read timeout",
			},
			"ip_addresses_info": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of IP addresses information of this dlb.",
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceDLBCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)
	body, err := newDLBPostBody(d)
	if err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create DLB of org " + orgid + " and vpc " + vpcid,
			Detail:   err.Error(),
		})
		return diags
	}
	//request user creation
	res, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersPost(authctx, orgid, vpcid).DlbPostBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
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

	return resourceDLBRead(ctx, d, m)
}

func resourceDLBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Id()
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)
	if isComposedResourceId(dlbid) {
		orgid, vpcid, dlbid = decomposeDlbId(d)
	}
	//request roles
	res, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdGet(authctx, orgid, vpcid, dlbid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
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
	d.SetId(dlbid)
	d.Set("org_id", orgid)
	d.Set("vpc_id", vpcid)
	return diags
}

func resourceDLBUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Id()
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)
	//check changes
	if isDLBChanged(ctx, d, m) {
		body := newDLBPatchBody(d)
		//request user creation
		_, httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdPatch(authctx, orgid, vpcid, dlbid).RequestBody(body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				defer httpr.Body.Close()
				b, _ := io.ReadAll(httpr.Body)
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
		return resourceDLBRead(ctx, d, m)
	}

	return diags
}

func resourceDLBDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	dlbid := d.Id()
	orgid := d.Get("org_id").(string)
	vpcid := d.Get("vpc_id").(string)
	authctx := getDLBAuthCtx(ctx, &pco)
	//perform request
	httpr, err := pco.dlbclient.DefaultApi.OrganizationsOrgIdVpcsVpcIdLoadbalancersDlbIdDelete(authctx, orgid, vpcid, dlbid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			defer httpr.Body.Close()
			b, _ := io.ReadAll(httpr.Body)
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

// Creates POST Body Object for request to creating a new DLB
func newDLBPostBody(d *schema.ResourceData) (*dlb.DlbPostBody, error) {
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
	if static_ips_disabled := d.Get("static_ips_disabled"); static_ips_disabled != nil {
		body.SetStaticIPsDisabled(static_ips_disabled.(bool))
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
		ssl_endpoints_map := newDlbPostBodySSLEndpointsMap(ssl_endpoints_set)
		ssl_endpoints_body, err := convertMap2DlbPostBodySslEndpoints(ssl_endpoints_map)
		if err != nil {
			return nil, err
		}
		body.SetSslEndpoints(ssl_endpoints_body)
	}
	return body, nil
}

// Creates a Patch Body Object to update a DLB
func newDLBPatchBody(d *schema.ResourceData) []map[string]interface{} {
	attributes := getDLBPatchWatchAttributes()
	body := make([]map[string]interface{}, len(attributes))
	op_replace := "replace"
	for i, attr := range attributes {
		camlAttr := strcase.ToLowerCamel(attr)
		item := make(map[string]interface{})
		if attr == "ssl_endpoints" {
			ssl_endpoints_set := d.Get(attr).(*schema.Set)
			ssl_endpoints_extract := newDlbPostBodySSLEndpointsMap(ssl_endpoints_set)
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

// Creates a, SSL Endpoint Requst Object from Terraform Schema Set
func newDlbPostBodySSLEndpointsMap(ssl_endpoints_set *schema.Set) []map[string]interface{} {
	ssl_endpoints_list := ssl_endpoints_set.List()
	ssl_endpoints_body := make([]map[string]interface{}, len(ssl_endpoints_list))
	public_key_field := "public_key"
	public_key_label_field := "public_key_label"
	private_key_field := "private_key"
	private_key_label_field := "private_key_label"
	client_cert_field := "client_cert"
	client_cert_label_field := "client_cert_label"
	revocation_list_field := "revocation_list"
	revocation_list_label_field := "revocation_list_label"
	verify_client_mode_field := "verify_client_mode"
	mappings_field := "mappings"
	mappings_input_uri_field := "input_uri"
	mappings_app_name_field := "app_name"
	mappings_app_uri_field := "app_uri"
	mappings_upstream_protocol_field := "upstream_protocol"
	for i, endpoint := range ssl_endpoints_list {
		endpoint_converted := endpoint.(map[string]interface{})
		endpoint_item := make(map[string]interface{})
		log.Println("****** PRINT SSL ENDPOINT BEFORE TRANSFORM ********")
		b2, _ := json.Marshal(endpoint_converted)
		log.Println(string(b2[:]))
		if val, ok := endpoint_converted[public_key_field]; ok {
			endpoint_item[strcase.ToLowerCamel(public_key_field)] = val.(string)
		}
		if val, ok := endpoint_converted[private_key_field]; ok {
			endpoint_item[strcase.ToLowerCamel(private_key_field)] = val.(string)
		}
		if val, ok := endpoint_converted[client_cert_field]; ok && len(val.(string)) > 0 {
			endpoint_item[strcase.ToLowerCamel(client_cert_field)] = val.(string)
		}
		if val, ok := endpoint_converted[revocation_list_field]; ok && len(val.(string)) > 0 {
			endpoint_item[strcase.ToLowerCamel(revocation_list_field)] = val.(string)
		}
		if val, ok := endpoint_converted[public_key_label_field]; ok {
			endpoint_item[strcase.ToLowerCamel(public_key_label_field)] = val.(string)
		}
		if val, ok := endpoint_converted[private_key_label_field]; ok {
			endpoint_item[strcase.ToLowerCamel(private_key_label_field)] = val.(string)
		}
		if val, ok := endpoint_converted[client_cert_label_field]; ok && len(val.(string)) > 0 {
			endpoint_item[strcase.ToLowerCamel(client_cert_label_field)] = val.(string)
		}
		if val, ok := endpoint_converted[revocation_list_label_field]; ok && len(val.(string)) > 0 {
			endpoint_item[strcase.ToLowerCamel(client_cert_label_field)] = val.(string)
		}
		if val, ok := endpoint_converted[verify_client_mode_field]; ok {
			endpoint_item[strcase.ToLowerCamel(verify_client_mode_field)] = val.(string)
		}
		if val, ok := endpoint_converted[mappings_field]; ok {
			mappings_set := val.(*schema.Set)
			mappings := mappings_set.List()
			mappings_body := make([]map[string]interface{}, len(mappings))
			for j, mapping := range mappings {
				mapping_converted := mapping.(map[string]interface{})
				mapping_item := make(map[string]interface{})
				if val, ok := mapping_converted[mappings_input_uri_field]; ok {
					mapping_item[strcase.ToLowerCamel(mappings_input_uri_field)] = val.(string)
				}
				if val, ok := mapping_converted[mappings_app_name_field]; ok {
					mapping_item[strcase.ToLowerCamel(mappings_app_name_field)] = val.(string)
				}
				if val, ok := mapping_converted[mappings_app_uri_field]; ok {
					mapping_item[strcase.ToLowerCamel(mappings_app_uri_field)] = val.(string)
				}
				if val, ok := mapping_converted[mappings_upstream_protocol_field]; ok {
					mapping_item[strcase.ToLowerCamel(mappings_upstream_protocol_field)] = val.(string)
				}
				mappings_body[j] = mapping_item
			}
			endpoint_item[strcase.ToLowerCamel(mappings_field)] = mappings_body
		}
		ssl_endpoints_body[i] = endpoint_item
	}
	log.Println("****** PRINT SSL ENDPOINTS BODY********")
	b, _ := json.Marshal(ssl_endpoints_body)
	log.Println(string(b[:]))
	return ssl_endpoints_body
}

func convertMap2DlbPostBodySslEndpoints(ssl_endpoints_array []map[string]interface{}) ([]dlb.DlbPostBodySslEndpoints, error) {
	list := make([]dlb.DlbPostBodySslEndpoints, len(ssl_endpoints_array))
	for i, endpoint := range ssl_endpoints_array {
		nullable_body := dlb.NewNullableDlbPostBodySslEndpoints(nil)
		b, e := json.Marshal(endpoint)
		if e != nil {
			return nil, e
		}
		nullable_body.UnmarshalJSON(b)
		list[i] = *nullable_body.Get()
	}
	return list, nil
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
	} else if new_lowercase == "started" && old_lowercase == "starting" {
		return true
	} else if new_lowercase == "stopped" && old_lowercase == "stopping" {
		return true
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

// Returns authentication context (includes authorization header)
func getDLBAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, dlb.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, dlb.ContextServerIndex, pco.server_index)
}

// Verifies if the source and its digest are valid
// uses CalcSha1Digest to calculate the digest against which it verifies validity
func verifyDLBDigest(source string, digest string) bool {
	return digest == CalcSha1Digest(source)
}

// Compares 2 states of DLB ssl_endpoints
// returns true if they are the same, false otherwise
func equalDLBSSLEndpoints(old, new interface{}) bool {
	old_set := old.(*schema.Set)
	old_list := old_set.List()
	new_set := new.(*schema.Set)
	new_list := new_set.List()

	sortAttr := []string{"private_key_label"}
	SortMapListAl(new_list, sortAttr)
	SortMapListAl(old_list, sortAttr)

	if len(new_list) != len(old_list) {
		return false
	}

	for i, val := range old_list {
		o := val.(map[string]interface{})
		n := new_list[i].(map[string]interface{})
		public_key := n["public_key"].(string)
		private_key := n["private_key"].(string)
		public_key_digest := o["public_key_digest"].(string)
		private_key_digest := o["private_key_digest"].(string)

		//compare certificates digest
		if !verifyDLBDigest(public_key, public_key_digest) || !verifyDLBDigest(private_key, private_key_digest) {
			return false
		}
		//compare client certificate digest
		if n["client_cert"] != nil && o["client_cert_digest"] != nil &&
			(!verifyDLBDigest(n["client_cert"].(string), o["client_cert_digest"].(string))) {
			return false
		}
		//compare revocation list digest
		if n["revocation_list"] != nil && o["revocation_list_digest"] != nil &&
			(!verifyDLBDigest(n["revocation_list"].(string), o["revocation_list_digest"].(string))) {
			return false
		}
		o_mapping_set := o["mappings"].(*schema.Set)
		n_mapping_set := n["mappings"].(*schema.Set)
		//compare mappings
		if !equalDLBSSLEndpointsMappings(o_mapping_set.List(), n_mapping_set.List()) {
			return false
		}
	}

	return true
}

// compares two SSL Endpoint Mappings
// returns true if they are equal, false otherwise
func equalDLBSSLEndpointsMappings(old, new []interface{}) bool {
	sortAttr := []string{"app_uri"}
	SortMapListAl(old, sortAttr)
	SortMapListAl(new, sortAttr)

	attributes := [...]string{
		"input_uri", "app_name", "app_uri",
	}

	if len(old) != len(new) {
		return false
	}

	for i, item := range old {
		old_mapping := item.(map[string]interface{})
		new_mapping := new[i].(map[string]interface{})
		//compare mapping attributes
		for _, attr := range attributes {
			if new_mapping[attr].(string) != old_mapping[attr].(string) {
				return false
			}
		}
	}
	return true
}

// Compares old and new values of allow list attribute
// returns true if they are the same, false otherwise
func equalDLBAllowList(old, new interface{}) bool {
	old_list := old.([]interface{})
	new_list := new.([]interface{})
	SortStrListAl(old_list)
	SortStrListAl(new_list)
	for i, item := range old_list {
		if new_list[i].(string) != item.(string) {
			return false
		}
	}
	return true
}

// returns true if the DLB key elements have been changed
func isDLBChanged(ctx context.Context, d *schema.ResourceData, m interface{}) bool {
	watchAttrs := getDLBPatchWatchAttributes()

	for _, attr := range watchAttrs {
		if attr == "ssl_endpoints" && !equalDLBSSLEndpoints(d.GetChange(attr)) {
			return true
		} else if attr == "ip_allowlist" {
			ip_allowlist := d.Get("ip_allowlist").([]interface{})
			if len(ip_allowlist) > 0 && !equalDLBAllowList(d.GetChange(attr)) {
				return true
			}
		} else if attr == "ip_whitelist" {
			ip_whitelist := d.Get("ip_whitelist").([]interface{})
			if len(ip_whitelist) > 0 && !equalDLBAllowList(d.GetChange(attr)) {
				return true
			}
		} else if d.HasChange(attr) {
			return true
		}
	}
	return false
}

func decomposeDlbId(d *schema.ResourceData) (string, string, string) {
	s := DecomposeResourceId(d.Id())
	return s[0], s[1], s[2]
}
