package anypoint

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	team "github.com/mulesoft-anypoint/anypoint-client-go/team"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Description: `
		Creates a ` + "`" + `team` + "`" + ` for your ` + "`" + `org` + "`" + `.
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
				Description: "The unique id of this team generated by the anypoint platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the team is defined.",
			},
			"parent_team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The team_id of the parent of this team.",
			},
			"team_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the team. Name is unique among teams within the organization.",
			},
			"team_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "internal",
				Description: `
				The type of the team. Internal teams are visible to all members of the organziation. 
				All internal teams of an organization are under the root internal team. 
				Private teams are internal teams but are only visible by maintainers/members of the team. 
				Shared teams are internal teams that can be mapped to external teams in other organizations where a trust relationship has been formed.
				Enum values are: internal, private and shared.
				`,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"internal", "external", "private", "shared", "legacy"}, true)),
			},
			"team_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the team. team_id is globally unique",
			},
			"ancestor_team_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Array of ancestor teams ids starting from either the internal or external root team down to this team's parent.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time the team was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time the team was last modified.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getTeamAuthCtx(ctx, &pco)
	body := newTeamPostBody(d)

	//request user creation
	res, httpr, err := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsPost(authctx, orgid).TeamPostBody(*body).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create team ",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(res.GetTeamId())
	return resourceTeamRead(ctx, d, m)
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	teamid := d.Id()
	orgid := d.Get("org_id").(string)
	if isComposedResourceId(teamid) {
		orgid, teamid = decomposeTeamId(d)
	}
	authctx := getTeamAuthCtx(ctx, &pco)

	//request roles
	res, httpr, err := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdGet(authctx, orgid, teamid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get team " + teamid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	team := flattenTeamData(&res)
	//save in data source schema
	if err := setTeamAttributesToResourceData(d, team); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set team " + teamid,
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(teamid)
	d.Set("org_id", orgid)

	return diags
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	teamid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getTeamAuthCtx(ctx, &pco)

	if d.HasChanges(getTeamPatchWatchAttributes()...) {
		body := newTeamPatchBody(d)
		//request user creation
		_, httpr, err := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdPatch(authctx, orgid, teamid).TeamPatchBody(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to patch team " + teamid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	if d.HasChanges(getTeamPutWatchAttributes()...) {
		body := newTeamPutBody(d)
		//request user creation
		_, httpr, err := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdParentPut(authctx, orgid, teamid).TeamPutBody(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil && httpr.StatusCode >= 400 {
				b, _ := io.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags := append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to move team " + teamid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceTeamRead(ctx, d, m)
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	teamid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getTeamAuthCtx(ctx, &pco)

	httpr, err := pco.teamclient.DefaultApi.OrganizationsOrgIdTeamsTeamIdDelete(authctx, orgid, teamid).Execute()
	if err != nil {
		var details string
		if httpr != nil && httpr.StatusCode >= 400 {
			b, _ := io.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete team " + teamid,
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

func newTeamPostBody(d *schema.ResourceData) *team.TeamPostBody {
	body := new(team.TeamPostBody)

	if parentTeamId := d.Get("parent_team_id"); parentTeamId != nil {
		body.SetParentTeamId(parentTeamId.(string))
	}
	if teamName := d.Get("team_name"); teamName != nil {
		body.SetTeamName(teamName.(string))
	}
	if teamType := d.Get("team_type"); teamType != nil {
		body.SetTeamType(teamType.(string))
	}

	return body
}

func newTeamPatchBody(d *schema.ResourceData) *team.TeamPatchBody {
	body := new(team.TeamPatchBody)
	if teamName := d.Get("team_name"); teamName != nil {
		body.SetTeamName(teamName.(string))
	}
	if teamType := d.Get("team_type"); teamType != nil {
		body.SetTeamType(teamType.(string))
	}

	return body
}

func newTeamPutBody(d *schema.ResourceData) *team.TeamPutBody {
	body := new(team.TeamPutBody)

	if parentTeamId := d.Get("parent_team_id"); parentTeamId != nil {
		body.SetParentTeamId(parentTeamId.(string))
	}

	return body
}

/*
List of attributes that requires patching the team
*/
func getTeamPatchWatchAttributes() []string {
	attributes := [...]string{
		"team_name", "team_type",
	}
	return attributes[:]
}

/*
List of attributes that requires to use put operation (to move team from one parent to another)
*/
func getTeamPutWatchAttributes() []string {
	attributes := [...]string{
		"parent_team_id",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getTeamAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, team.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, team.ContextServerIndex, pco.server_index)
}

func decomposeTeamId(d *schema.ResourceData, separator ...string) (string, string) {
	s := DecomposeResourceId(d.Id(), separator...)
	return s[0], s[1]
}
