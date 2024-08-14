package anypoint

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	application_manager_v2 "github.com/mulesoft-anypoint/anypoint-client-go/application_manager_v2"
)

var ReplicasReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The unique id of the mule app replica.",
		},
		"state": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The current state of the replica.",
		},
		"deployment_location": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The node id in which the replica is deployed.",
		},
		"current_deployment_version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The version deployed in the replica.",
		},
		"reason": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "In case of an error, it should provide information about the root cause.",
		},
	},
}

var ApplicationRefReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"group_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The groupId of the application.",
		},
		"artifact_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The artifactId of the application.",
		},
		"version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The version of the application.",
		},
		"packaging": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The packaging of the application.",
		},
	},
}

var ApplicationConfigPropsReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"application_name": {
			Type:        schema.TypeString,
			Description: "The application name",
			Computed:    true,
		},
		"properties": {
			Type:        schema.TypeMap,
			Description: "The mule application properties.",
			Computed:    true,
		},
		"secure_properties": {
			Type:        schema.TypeMap,
			Description: "The mule application secured properties.",
			Computed:    true,
		},
	},
}

var ApplicationConfigLoggingReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"artifact_name": {
			Type:        schema.TypeString,
			Description: "The application name.",
			Computed:    true,
		},
		"scope_logging_configurations": {
			Type:        schema.TypeList,
			Description: "Additional log levels and categories to include in logs.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"scope": {
						Type:        schema.TypeString,
						Description: "The logging package scope",
					},
					"log_level": {
						Type:        schema.TypeString,
						Description: "The application log level: INFO / DEBUG / WARNING / ERROR / FATAL",
						// ValidateDiagFunc: validation.ToDiagFunc(
						// 	validation.StringInSlice([]string{"INFO", "DEBUG", "WARNING", "ERROR", "FATAL"}, false),
						// ),
					},
				},
			},
		},
	},
}

var ApplicationConfigSchedulingReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"application_name": {
			Type:        schema.TypeString,
			Description: "The mule application name.",
			Computed:    true,
		},
		"schedulers": {
			Type:        schema.TypeList,
			Description: "The mule app schedulers details",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Description: "The scheduler name",
						Computed:    true,
					},
					"type": {
						Type:        schema.TypeString,
						Description: "The scheduler type",
					},
					"flow_name": {
						Type:        schema.TypeString,
						Description: "The scheduler flow name",
						Computed:    true,
					},
					"enabled": {
						Type:        schema.TypeBool,
						Description: "Whether the scheduler is enabled or not.",
						Computed:    true,
					},
					"time_unit": {
						Type:        schema.TypeString,
						Description: "The scheduler's time unit.",
						Computed:    true,
						ValidateDiagFunc: validation.ToDiagFunc(
							validation.StringInSlice(
								[]string{"NANOSECONDS", "MICROSECONDS", "MILLISECONDS", "SECONDS", "MINUTES", "HOURS", "DAYS"},
								false,
							),
						),
					},
					"frequency": {
						Type:        schema.TypeString,
						Description: "The scheduler's frequency",
						Computed:    true,
					},
					"start_delay": {
						Type:        schema.TypeString,
						Description: "The scheduler's start delay",
						Computed:    true,
					},
					"expression": {
						Type:        schema.TypeString,
						Description: "The scheduler's cron expression",
						Computed:    true,
					},
					"time_zone": {
						Type:        schema.TypeString,
						Description: "The scheduler's time zone",
						Computed:    true,
					},
				},
			},
		},
	},
}

var ApplicationConfigReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"mule_agent_app_props_service": {
			Type:        schema.TypeList,
			Description: "The mule app properties",
			Elem:        ApplicationConfigPropsReadOnlyDefinition,
		},
		"mule_agent_logging_service": {
			Type:        schema.TypeList,
			Description: "The mule app logging props",
			Elem:        ApplicationConfigLoggingReadOnlyDefinition,
		},
		"mule_agent_scheduling_service": {
			Type:        schema.TypeList,
			Description: "The mule app scheduling",
			Elem:        ApplicationConfigSchedulingReadOnlyDefinition,
		},
	},
}

var ApplicationReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the application.",
		},
		"desired_state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The desired state of the application.",
			// ValidateDiagFunc: validation.ToDiagFunc(
			// 	validation.StringInSlice(
			// 		[]string{
			// 			"PARTIALLY_STARTED", "DEPLOYMENT_FAILED", "STARTING", "STARTED", "STOPPING",
			// 			"STOPPED", "UNDEPLOYING", "UNDEPLOYED", "UPDATED", "APPLIED", "APPLYING", "FAILED", "DELETED",
			// 		},
			// 		false,
			// 	),
			// ),
		},
		"ref": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The desired state of the application.",
			Elem:        ApplicationRefReadOnlyDefinition,
		},
		"configuration": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The configuration of the application.",
			Elem:        ApplicationConfigReadOnlyDefinition,
		},
		"vcores": {
			Type:        schema.TypeFloat,
			Computed:    true,
			Description: "The allocated virtual cores.",
		},
	},
}

var TargetDeploymentSettingsHttpReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"inbound_public_url": {
			Type:        schema.TypeString,
			Description: "The inbound public url",
			Computed:    true,
		},
		"inbound_path_rewrite": {
			Type:        schema.TypeString,
			Description: "The inbound path rewrite",
			Computed:    true,
		},
		"inbound_last_mile_security": {
			Type:        schema.TypeBool,
			Description: "Last-mile security means that the connection between ingress and the actual Mule app will be HTTPS.",
			Computed:    true,
		},
		"inbound_forward_ssl_session": {
			Type:        schema.TypeBool,
			Description: "Whether to forward the ssl session",
			Computed:    true,
		},
		"inbound_internal_url": {
			Type:        schema.TypeString,
			Description: "The inbound internal url",
			Computed:    true,
		},
		"inbound_unique_id": {
			Type:        schema.TypeString,
			Description: "The inbound unique id",
		},
	},
}

var TargetDeploymentSettingsRuntimeReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"version": {
			Type: schema.TypeString,
			Description: `
			On deployment operations it can be set to:
				- a full image version with tag (i.e "4.6.0:40e-java17"),
				- a base version with a partial tag not indicating the java version (i.e. "4.6.0:40")
				- or only a base version (i.e. "4.6.0").
			Defaults to the latest image version.
			This field has precedence over the legacy 'target.deploymentSettings.runtimeVersion'
			`,
			Computed: true,
		},
		"release_channel": {
			Type: schema.TypeString,
			Description: `
			On deployment operations it can be set to one of:
				- "LTS"
				- "EDGE"
				- "LEGACY".
			Defaults to "EDGE". This field has precedence over the legacy 'target.deploymentSettings.runtimeReleaseChannel'.
			Learn more on release channels at https://docs.mulesoft.com/release-notes/mule-runtime/lts-edge-release-cadence.
			`,
			Computed: true,
		},
		"java": {
			Type: schema.TypeString,
			Description: `
			On deployment operations it can be set to one of:
				- "8"
				- "17"
			Defaults to "8".
			Learn more about Java support at https://docs.mulesoft.com/general/java-support.
			`,
			Computed: true,
		},
	},
}

var TargetDeploymentSettingsAutoscalingReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Enables or disables the Autoscaling feature. The possible values are: true or false.",
			Computed:    true,
		},
		"min_replicas": {
			Type:        schema.TypeInt,
			Description: "Set the minimum amount of replicas for your deployment. The minimum accepted value is 1. The maximum is 3.",
			Computed:    true,
		},
		"max_replicas": {
			Type:        schema.TypeInt,
			Description: "Set the maximum amount of replicas your application can scale to. The minimum accepted value is 2. The maximum is 32.",
			Computed:    true,
		},
	},
}

var TargetDeploymentSettingsResourcesReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu_limit": {
			Type:        schema.TypeString,
			Description: "The CPU limit",
			Computed:    true,
		},
		"cpu_reserved": {
			Type:        schema.TypeString,
			Description: "The CPU reserved",
			Computed:    true,
		},
		"memory_limit": {
			Type:        schema.TypeString,
			Description: "The memory limit",
			Computed:    true,
		},
		"memory_reserved": {
			Type:        schema.TypeString,
			Description: "The memory reserved",
			Computed:    true,
		},
		"storage_limit": {
			Type:        schema.TypeString,
			Description: "The storage limit",
			Computed:    true,
		},
		"storage_reserved": {
			Type:        schema.TypeString,
			Description: "The storage reserved",
			Computed:    true,
		},
	},
}

var TargetDeploymentSettingsSidecarsReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"anypoint_monitoring_image": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"anypoint_monitoring_resources_cpu_limit": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"anypoint_monitoring_resources_cpu_reserved": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"anypoint_monitoring_resources_memory_limit": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"anypoint_monitoring_resources_memory_reserved": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

var TargetDeploymentSettingsReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"clustered": {
			Type:        schema.TypeBool,
			Description: "Whether the application is deployed in clustered mode.",
			Computed:    true,
		},
		"enforce_deploying_replicas_across_nodes": {
			Type:        schema.TypeBool,
			Description: "If true, forces the deployment of replicas across the RTF cluster. This option only available for Runtime Fabrics.",
			Computed:    true,
		},
		"http": {
			Type:        schema.TypeList,
			Description: "The details about http inbound or outbound configuration",
			Computed:    true,
			Elem:        TargetDeploymentSettingsHttpReadOnlyDefinition,
		},
		"jvm_args": {
			Type:        schema.TypeString,
			Description: "The java virtual machine arguments",
			Computed:    true,
		},
		"runtime": {
			Type:        schema.TypeList,
			Description: "The Mule app runtime version info.",
			Computed:    true,
			Elem:        TargetDeploymentSettingsRuntimeReadOnlyDefinition,
		},
		"autoscaling": {
			Type: schema.TypeList,
			Description: `
			Use this object to provide CPU Based Horizontal Autoscaling configuration on deployment and redeployment operations. This object is optional.
			If Autoscaling is disabled and the fields "minReplicas" and "maxReplicas" are provided, they must match the value of "target.replicas" field.
			Learn more about Autoscaling at https://docs.mulesoft.com/cloudhub-2/ch2-configure-horizontal-autoscaling.
			`,
			Computed: true,
			Elem:     TargetDeploymentSettingsAutoscalingReadOnlyDefinition,
		},
		"update_strategy": {
			Type:        schema.TypeString,
			Description: "The mule app update strategy: rolling or replace",
			Computed:    true,
		},
		"resources": {
			Type:        schema.TypeList,
			Description: "The mule app allocated resources",
			Elem:        TargetDeploymentSettingsResourcesReadOnlyDefinition,
			Computed:    true,
		},
		"last_mile_security": {
			Type:        schema.TypeBool,
			Description: "Whether last mile security is active",
			Computed:    true,
		},
		"disable_am_log_forwarding": {
			Type:        schema.TypeBool,
			Description: "Whether log forwarding is disabled.",
			Computed:    true,
		},
		"persistent_object_store": {
			Type:        schema.TypeBool,
			Description: "Whether persistent object store is enabled.",
			Computed:    true,
		},
		"anypoint_monitoring_scope": {
			Type:        schema.TypeString,
			Description: "The anypoint moniroting scope",
		},
		"sidecars": {
			Type:        schema.TypeList,
			Description: "The mule app sidecars.",
			Elem:        TargetDeploymentSettingsSidecarsReadOnlyDefinition,
			Computed:    true,
		},
		"forward_ssl_session": {
			Type:        schema.TypeBool,
			Description: "Whether the ssl session is forwarded to the mule app.",
			Computed:    true,
		},
		"disable_external_log_forwarding": {
			Type:        schema.TypeBool,
			Description: "Whether the log forwarding is disabled.",
			Computed:    true,
		},
		"tracing_enabled": {
			Type:        schema.TypeBool,
			Description: "Whether the log tracing is enabled.",
			Computed:    true,
		},
		"generate_default_public_url": {
			Type:        schema.TypeBool,
			Description: "Whether default public url should be generated",
			Computed:    true,
		},
	},
}

var TargetReadOnlyDefinition = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"provider": {
			Type:        schema.TypeString,
			Description: "The cloud provider the target belongs to.",
			Computed:    true,
		},
		"target_id": {
			Type:        schema.TypeString,
			Description: "The unique identifier of the target.",
			Computed:    true,
		},
		"deployment_settings": {
			Type:        schema.TypeList,
			Description: "The settings of the target for the deployment to perform.",
			Elem:        TargetDeploymentSettingsReadOnlyDefinition,
		},
		"replicas": {
			Type:        schema.TypeInt,
			Description: "The number of replicas",
			Computed:    true,
		},
	},
}

func dataSourceAppDeploymentV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppDeploymentV2Read,
		Description: `
		Reads a specific ` + "`" + `Deployment` + "`" + `.
		This only works for Cloudhub V2 and Runtime Fabrics Apps.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique id of the mule app deployment in the platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organization where the mule app is deployed.",
			},
			"env_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment where mule app is deployed.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the deployed mule app.",
			},
			"creation_date": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The creation date of the mule app.",
			},
			"last_modified_date": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The last modification date of the mule app.",
			},
			"desired_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployment desired version of the mule app.",
			},
			"replicas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Data of the mule app replicas",
				Elem:        ReplicasReadOnlyDefinition,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data of the mule app replicas",
			},
			"application": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The details of the application to deploy",
				Elem:        ApplicationReadOnlyDefinition,
			},
			"target": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The details of the target to perform the deployment on.",
				Elem:        TargetReadOnlyDefinition,
			},
			"last_successful_version": {
				Type:        schema.TypeString,
				Description: "The last successfully deployed version",
				Computed:    true,
			},
		},
	}
}

func dataSourceAppDeploymentV2Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	id := d.Get("id").(string)
	orgid := d.Get("org_id").(string)
	envid := d.Get("env_id").(string)
	authctx := getAppDeploymentV2AuthCtx(ctx, &pco)
	//execut request
	res, httpr, err := pco.appmanagerclient.DefaultApi.GetDeploymentById(authctx, orgid, envid, id).Execute()
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
			Summary:  "Unable to get deployment for org " + orgid + " and env " + envid + " with id " + id,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	data := flattenAppDeployment(res)
	if err := setAppDeploymentV2AttributesToResourceData(d, data); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set API Manager instance details attributes",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(res.GetId())
	return diags
}

func flattenAppDeployment(deployment *application_manager_v2.Deployment) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := deployment.GetNameOk(); ok {
		item["name"] = *val
	}
	if val, ok := deployment.GetCreationDateOk(); ok {
		item["creation_date"] = *val
	}
	if val, ok := deployment.GetLastModifiedDateOk(); ok {
		item["last_modified_date"] = *val
	}
	if val, ok := deployment.GetDesiredVersionOk(); ok {
		item["desired_version"] = *val
	}
	if val, ok := deployment.GetReplicasOk(); ok {
		item["replicas"] = flattenAppDeploymentReplicas(val)
	}
	if val, ok := deployment.GetStatusOk(); ok {
		item["status"] = *val
	}
	if val, ok := deployment.GetApplicationOk(); ok {
		item["application"] = flattenAppDeploymentApplication(val)
	}
	return item
}

func flattenAppDeploymentReplicas(replicas []application_manager_v2.Replicas) []interface{} {
	if len(replicas) > 0 {
		res := make([]interface{}, len(replicas))
		for i, replica := range replicas {
			res[i] = flattenAppDeploymentReplica(&replica)
		}
		return res
	}
	return make([]interface{}, 0)
}

func flattenAppDeploymentReplica(replica *application_manager_v2.Replicas) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := replica.GetIdOk(); ok {
		item["id"] = *val
	}
	if val, ok := replica.GetStateOk(); ok {
		item["state"] = *val
	}
	if val, ok := replica.GetDeploymentLocationOk(); ok {
		item["deployment_location"] = *val
	}
	if val, ok := replica.GetCurrentDeploymentVersionOk(); ok {
		item["current_deployment_version"] = *val
	}
	if val, ok := replica.GetReasonOk(); ok {
		item["reason"] = *val
	}
	return item
}

func flattenAppDeploymentApplication(application *application_manager_v2.Application) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := application.GetStatusOk(); ok {
		item["status"] = *val
	}
	if val, ok := application.GetDesiredStateOk(); ok {
		item["desired_state"] = *val
	}
	if ref, ok := application.GetRefOk(); ok {
		item["ref"] = []interface{}{flattenAppDeploymentRef(ref)}
	}
	if config, ok := application.GetConfigurationOk(); ok {
		item["configuration"] = []interface{}{flattenAppDeploymentConfig(config)}
	}
	return item
}

func flattenAppDeploymentRef(ref *application_manager_v2.Ref) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := ref.GetGroupIdOk(); ok {
		item["group_id"] = *val
	}
	if val, ok := ref.GetArtifactIdOk(); ok {
		item["artifact_id"] = *val
	}
	if val, ok := ref.GetVersionOk(); ok {
		item["version"] = *val
	}
	if val, ok := ref.GetPackagingOk(); ok {
		item["packaging"] = *val
	}
	return item
}

func flattenAppDeploymentConfig(config *application_manager_v2.AppConfiguration) map[string]interface{} {
	item := make(map[string]interface{})
	if srv, ok := config.GetMuleAgentApplicationPropertiesServiceOk(); ok {
		item["mule_agent_app_props_service"] = []interface{}{flattenAppDeploymentConfigMAAPS(srv)}
	}
	if srv, ok := config.GetMuleAgentLoggingServiceOk(); ok {
		item["mule_agent_logging_service"] = []interface{}{flattenAppDeploymentConfigMALS(srv)}
	}
	if srv, ok := config.GetMuleAgentSchedulingServiceOk(); ok {
		item["mule_agent_scheduling_service"] = []interface{}{flattenAppDeploymentConfigMASS(srv)}
	}
	return item
}

func flattenAppDeploymentConfigMAAPS(service *application_manager_v2.MuleAgentAppPropService) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := service.GetApplicationNameOk(); ok {
		item["application_name"] = *val
	}
	if val, ok := service.GetPropertiesOk(); ok {
		item["properties"] = val
	}
	if val, ok := service.GetSecurePropertiesOk(); ok {
		item["secure_properties"] = val
	}
	return item
}

func flattenAppDeploymentConfigMALS(service *application_manager_v2.MuleAgentLoggingService) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := service.GetArtifactNameOk(); ok {
		item["artifact_name"] = *val
	}
	if scope_logging_conf, ok := service.GetScopeLoggingConfigurationsOk(); ok {
		res := make([]interface{}, len(scope_logging_conf))
		for i, cfg := range scope_logging_conf {
			d := make(map[string]interface{})
			if val, ok := cfg.GetScopeOk(); ok {
				d["scope"] = *val
			}
			if val, ok := cfg.GetLogLevelOk(); ok {
				d["log_level"] = *val
			}
			res[i] = d
		}
		item["scope_logging_configurations"] = res
	}
	return item
}

func flattenAppDeploymentConfigMASS(service *application_manager_v2.MuleAgentSchedulingService) map[string]interface{} {
	item := make(map[string]interface{})
	if val, ok := service.GetApplicationNameOk(); ok {
		item["application_name"] = *val
	}
	if schedulers, ok := service.GetSchedulersOk(); ok {
		res := make([]interface{}, len(schedulers))
		for i, scheduler := range schedulers {
			d := make(map[string]interface{})
			if val, ok := scheduler.GetNameOk(); ok {
				d["name"] = *val
			}
			if val, ok := scheduler.GetTypeOk(); ok {
				d["type"] = *val
			}
			if val, ok := scheduler.GetFlowNameOk(); ok {
				d["flow_name"] = *val
			}
			if val, ok := scheduler.GetEnabledOk(); ok {
				d["enabled"] = *val
			}
			if val, ok := scheduler.GetTimeUnitOk(); ok {
				d["time_unit"] = *val
			}
			if val, ok := scheduler.GetFrequencyOk(); ok {
				d["frequency"] = *val
			}
			if val, ok := scheduler.GetStartDelayOk(); ok {
				d["start_delay"] = *val
			}
			if val, ok := scheduler.GetExpressionOk(); ok {
				d["expression"] = *val
			}
			if val, ok := scheduler.GetTimeZoneOk(); ok {
				d["time_zone"] = *val
			}
			res[i] = d
		}
		item["schedulers"] = res
	}
	return item
}

func setAppDeploymentV2AttributesToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	attributes := getAppDeploymentV2Attributes()
	if data != nil {
		for _, attr := range attributes {
			if val, ok := data[attr]; ok {
				if err := d.Set(attr, val); err != nil {
					return fmt.Errorf("unable to set app deployment attribute %s\n\tdetails: %s", attr, err)
				}
			}
		}
	}
	return nil
}

func getAppDeploymentV2Attributes() []string {
	attributes := [...]string{
		"name", "creation_date", "last_modified_date", "desired_version",
		"replicas", "status", "application", "target", "last_successful_version",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getAppDeploymentV2AuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, application_manager_v2.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, application_manager_v2.ContextServerIndex, pco.server_index)
}
