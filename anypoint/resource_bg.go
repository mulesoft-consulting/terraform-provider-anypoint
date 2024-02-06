package anypoint

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	org "github.com/mulesoft-anypoint/anypoint-client-go/org"
)

func resourceBG() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBGCreate,
		ReadContext:   resourceBGRead,
		UpdateContext: resourceBGUpdate,
		DeleteContext: resourceBGDelete,
		Description: `
		Creates a business group (org).
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
				Description: "This organization's unique id generated by the anypoint plaform",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of this organization.",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user id of the owner of this organization.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when this organization was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when this organization was updated.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization client id.",
			},
			"idprovider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identity provider if of this organization",
			},
			"is_federated": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether this organization is federated.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false")
				},
			},
			"parent_organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The immediate parent organization id of this organization.",
			},
			"parent_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of ancestor organizations.",
			},
			"sub_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of descendant organizations.",
			},
			"tenant_organization_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of tenant organizations",
			},
			"mfa_required": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Whether MFA is enforced in this organization",
			},
			"is_automatic_admin_promotion_exempt": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the admin promotion exemption is enabled on this organization",
			},
			"domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization's domain",
			},
			"is_master": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization is the master org.",
			},
			"subscription_category": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The anypoint platform subscription category",
			},
			"subscription_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The anypoint platform subscription type.",
			},
			"subscription_expiration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The anypoint platform subscription expiration date.",
			},
			"properties": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organiztion's general properties.",
			},
			"environments": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The organization's list of environments",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment unique id.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment name",
						},
						"organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment's organization id.",
						},
						"is_production": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this environment is a production environment.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the environment (e.g sandbox or production)",
						},
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The environment's client id",
						},
					},
				},
			},
			"entitlements_createenvironments": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this organization can have additional environments.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false")
				},
			},
			"entitlements_globaldeployment": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this organization can have global deployments.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false")
				},
			},
			"entitlements_createsuborgs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether this organization can create sub organizations (descendants).",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true")
				},
			},
			"entitlements_hybridenabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization has hybrid enabled.",
			},
			"entitlements_hybridinsight": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization has hybrid insight.",
			},
			"entitlements_hybridautodiscoverproperties": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization has hybrid auto-discovery properties enabled",
			},
			"entitlements_vcoresproduction_assigned": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     0,
				Description: "The number of production vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0")
				},
			},
			"entitlements_vcoresproduction_reassigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of production vcores reassigned to this organization.",
			},
			"entitlements_vcoressandbox_assigned": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     0,
				Description: "The number of sandbox vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0")
				},
			},
			"entitlements_vcoressandbox_reassigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of sandbox vcores reassigned to this organization.",
			},
			"entitlements_vcoresdesign_assigned": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     0,
				Description: "The number of design vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0")
				},
			},
			"entitlements_vcoresdesign_reassigned": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of design vcores reassigned to this organization.",
			},
			"entitlements_staticips_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of static IPs assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0")
				},
			},
			"entitlements_staticips_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of static IPs reassigned to this organization.",
			},
			"entitlements_vpcs_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of VPCs assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0")
				},
			},
			"entitlements_vpcs_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of VPCs reassigned to this organization.",
			},
			"entitlements_vpns_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of VPNs assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0")
				},
			},
			"entitlements_vpns_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of VPNs reassigned to this organization.",
			},
			"entitlements_workerloggingoverride_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the loggin override on workers is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_mqmessages_base": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     50000000,
				Description: "The number of basic MQ messages assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "50000000") // default value of integers if not set is 50000000
				},
			},
			"entitlements_mqmessages_addon": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of MQ messages addons assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_mqrequests_base": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100000000,
				Description: "The number of MQ requests base assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "100000000") // default value of integers if not set is 100000000
				},
			},
			"entitlements_mqrequests_addon": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of MQ requests addon assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_objectstorerequestunits_base": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of object store requests unists base for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_objectstorerequestunits_addon": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of object store requests units addon for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_objectstorekeys_base": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of object store keys base for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_objectstorekeys_addon": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of object store keys addon for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_mqadvancedfeatures_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the Anypoint MQ advanced features are enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_gateways_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of gateways assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_designcenter_api": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether te design center api is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_designcenter_mozart": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the design center mozart is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_partnersproduction_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of partners production vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_partnerssandbox_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of partners sandbox vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_tradingpartnersproduction_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of traded partners production vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_tradingpartnerssandbox_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of traded partners sandbox vcores assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_loadbalancer_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of dedicated load balancers (DLB) assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_loadbalancer_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of dedicated load balancers (DLB) reassigned to this organization.",
			},
			"entitlements_externalidentity": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether an external identity provider (IDP) was assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_autoscaling": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether autoscaling is enabled for this organization",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_armalerts": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether arm alerts are enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_apis_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "whether APIs are enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_apimonitoring_schedules": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "The number of api monitoring schedules for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "5") // default value of integers if not set is 0
				},
			},
			"entitlements_apicommunitymanager_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether api community manager is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_monitoringcenter_productsku": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "The number of monitoring center products sku for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "3") // default value of integers if not set is 0
				},
			},
			"entitlements_apiquery_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether api queries are enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_apiquery_productsku": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of api query product sku for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "1") // default value of integers if not set is 0
				},
			},
			"entitlements_apiqueryc360_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether api query C360 is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_anggovernance_level": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integers if not set is 0
				},
			},
			"entitlements_crowd_hideapimanagerdesigner": {
				Type:     schema.TypeBool,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_crowd_hideformerapiplatform": {
				Type:     schema.TypeBool,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_crowd_environments": {
				Type:     schema.TypeBool,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_cam_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether cam is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_exchange2_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether exchange v2 is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_crowdselfservicemigration_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether crow self service migration is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_kpidashboard_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether KPI dashboard is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_pcf": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether PCF is included for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_appviz": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the app vizualize if enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "flase") // default value of bool if not set is false
				},
			},
			"entitlements_runtimefabric": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether Runtime Fabrics (RTF) is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true")
				},
			},
			"entitlements_anypointsecuritytokenization_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "whether Anypoint securirty tokenization is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_anypointsecurityedgepolicies_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether Anypoint security edge policies is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_runtimefabriccloud_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether Runtime Fabrics (RTF) is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "true") // default value of bool if not set is false
				},
			},
			"entitlements_servicemesh_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether Service Mesh is enabled for this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "false") // default value of bool if not set is false
				},
			},
			"entitlements_messaging_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of messaging assigned to this organization.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "1") // default value of integers if not set is 0
				},
			},
			"entitlements_workerclouds_assigned": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of worker clouds assigned to this organization",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "1") // default value of integers if not set is 0
				},
			},
			"entitlements_workerclouds_reassigned": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of worker clouds reassigned to this organization",
			},
			"owner_created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the organization owner creation date",
			},
			"owner_updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner update date.",
			},
			"owner_organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner's organization id.",
			},
			"owner_firstname": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's firstname",
			},
			"owner_lastname": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's lastname.",
			},
			"owner_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's email.",
			},
			"owner_phonenumber": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The organization owner's phone number.",
			},
			"owner_username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner username.",
			},
			"owner_idprovider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner identity provider id.",
			},
			"owner_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the organization owner account is enabled.",
			},
			"owner_deleted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the organization owner account is deleted.",
			},
			"owner_lastlogin": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time the organization owner logged in.",
			},
			"owner_mfaverification_excluded": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the organization owner MFA verification is excluded.",
			},
			"owner_mfaverifiers_configured": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner MFA verification configuration",
			},
			"owner_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization owner account type.",
			},
			"session_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The organization's session timeout",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return DiffSuppressFunc4OptionalPrimitives(k, old, new, d, "0") // default value of integeres if not set is 0
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceBGCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	authctx := getBGAuthCtx(ctx, &pco)
	body := newBGPostBody(d)
	//perform request
	res, httpr, err := pco.orgclient.DefaultApi.OrganizationsPost(authctx).BGPostReqBody(*body).Execute()
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
			Summary:  "Unable to Create Business Group",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(res.GetId())
	return resourceBGRead(ctx, d, m)
}

func resourceBGRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Id()
	authctx := getBGAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdGet(authctx, orgid).Execute()
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
			Summary:  "Unable to read business group " + orgid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process response data
	orginstance := flattenBGData(&res)
	if err := setBGCoreAttributesToResourceData(d, orginstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Business Group",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(orgid)
	return diags
}

func resourceBGUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Id()
	authctx := getBGAuthCtx(ctx, &pco)
	//check for updates
	if d.HasChanges(getBGUpdatableAttributes()...) {
		body := newBGPutBody(d)
		_, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdPut(authctx, orgid).BGPutReqBody(*body).Execute()
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
				Summary:  "Unable to update business group " + orgid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceBGRead(ctx, d, m)
	}
	return diags
}

func resourceBGDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Id()
	authctx := getBGAuthCtx(ctx, &pco)
	//perform request
	_, httpr, err := pco.orgclient.DefaultApi.OrganizationsOrgIdDelete(authctx, orgid).Execute()
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
			Summary:  "Unable to Delete Business Group",
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

/*
 * Creates body for B.G POST request
 */
func newBGPostBody(d *schema.ResourceData) *org.BGPostReqBody {
	body := org.NewBGPostReqBodyWithDefaults()

	body.SetName(d.Get("name").(string))
	body.SetOwnerId(d.Get("owner_id").(string))
	body.SetParentOrganizationId((d.Get("parent_organization_id").(string)))
	body.SetEntitlements(*newEntitlementsFromD(d))

	return body
}

/*
 * Creates body for B.G PUT request
 */
func newBGPutBody(d *schema.ResourceData) *org.BGPutReqBody {
	body := org.NewBGPutReqBodyWithDefaults()
	body.SetName(d.Get("name").(string))
	body.SetOwnerId(d.Get("owner_id").(string))
	body.SetEntitlements(*newEntitlementsFromD(d))
	body.SetSessionTimeout(int32(d.Get("session_timeout").(int)))

	return body
}

/*
 * Creates Entitlements from Resource Data Schema
 */
func newEntitlementsFromD(d *schema.ResourceData) *org.EntitlementsCore {
	loadbalancer := org.NewLoadBalancerWithDefaults()
	loadbalancer.SetAssigned(int32(d.Get("entitlements_loadbalancer_assigned").(int)))
	staticips := org.NewStaticIpsWithDefaults()
	staticips.SetAssigned(int32(d.Get("entitlements_staticips_assigned").(int)))
	vcoresandbox := org.NewVCoresSandboxWithDefaults()
	vcoresandbox.SetAssigned(float32(d.Get("entitlements_vcoressandbox_assigned").(float64)))
	vcoredesign := org.NewVCoresDesignWithDefaults()
	vcoredesign.SetAssigned(float32(d.Get("entitlements_vcoresdesign_assigned").(float64)))
	vpns := org.NewVpnsWithDefaults()
	vpns.SetAssigned(int32(d.Get("entitlements_vpns_assigned").(int)))
	vpcs := org.NewVpcsWithDefaults()
	vpcs.SetAssigned(int32(d.Get("entitlements_vpcs_assigned").(int)))
	vcoreprod := org.NewVCoresProductionWithDefaults()
	vcoreprod.SetAssigned(float32(d.Get("entitlements_vcoresproduction_assigned").(float64)))
	entitlements := org.NewEntitlementsCore(
		d.Get("entitlements_globaldeployment").(bool),
		d.Get("entitlements_createenvironments").(bool),
		d.Get("entitlements_createsuborgs").(bool),
		*loadbalancer,
		*staticips,
		*vcoredesign,
		*vcoreprod,
		*vcoresandbox,
		*vpcs,
		*vpns,
	)

	return entitlements
}

/*
 * Returns authentication context (includes authorization header)
 */
func getBGAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, org.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, org.ContextServerIndex, pco.server_index)
}
