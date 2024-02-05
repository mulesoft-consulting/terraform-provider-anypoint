package anypoint

import (
	"context"
	"io"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-anypoint/anypoint-client-go/user"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Description: `
		Creates a ` + "`" + `user` + "`" + ` for your org. 

**N.B:** you can use a username only once even after it's deleted.
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
				Description: "The unique id of this user generated by the anypoint platform.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The master organization id where the user is defined.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username of this user.",
			},
			"first_name": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The firstname of this user.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The lastname of this user.",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The email of this user.",
			},
			"phone_number": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The phone number of this user.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The password of this user.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The master organization id where the user is defined.",
			},
			"idprovider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identity provider id",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this user is enabled",
			},
			"deleted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this user is deleted",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when the user was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time this user was updated.",
			},
			"last_login": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time this user logged in.",
			},
			"mfa_verifiers_configured": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The MFA configured for this user.",
			},
			"mfa_verification_excluded": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether MFA verification is excluded for this user",
			},
			"is_federated": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this user is federated.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of user.",
			},
			"organization_preferences": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The preferences of the user within the organization.",
			},
			"member_of_organizations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Description: "The user's list of organizations membership",
			},
			"contributor_of_organizations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Description: "The list of organizations this user has contributed to.",
			},
			"organization": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The organization information",
			},
			"properties": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's properties.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	username := d.Get("username").(string)
	authctx := getUserAuthCtx(ctx, &pco)
	//prepare request body
	body := newUserPostBody(d)
	//perform request
	res, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersPost(authctx, orgid).UserPostBody(*body).Execute()
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
			Summary:  "Unable to create user " + username,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId(res.GetId())
	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Id()
	orgid := d.Get("org_id").(string)
	if isComposedResourceId(userid) {
		orgid, userid = decomposeUserId(d)
	}
	authctx := getUserAuthCtx(ctx, &pco)
	//perform request
	res, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdGet(authctx, orgid, userid).Execute()
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
			Summary:  "Unable to Get User " + userid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	//process data
	user := flattenUserData(&res)
	//save in data source schema
	if err := setUserAttributesToResourceData(d, user); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set User",
			Detail:   err.Error(),
		})
		return diags
	}
	//set identifiers params
	d.SetId(userid)
	d.Set("org_id", orgid)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getUserAuthCtx(ctx, &pco)
	//check for updates
	if d.HasChanges(getUserWatchAttributes()...) {
		body := newUserPutBody(d)
		//request user creation
		_, httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdPut(authctx, orgid, userid).UserPutBody(*body).Execute()
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
				Summary:  "Unable to update user " + userid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
		d.Set("last_updated", time.Now().Format(time.RFC850))
		return resourceUserRead(ctx, d, m)
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	userid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getUserAuthCtx(ctx, &pco)
	//perform request
	httpr, err := pco.userclient.DefaultApi.OrganizationsOrgIdUsersUserIdDelete(authctx, orgid, userid).Execute()
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
			Summary:  "Unable to delete user " + userid,
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

func newUserPostBody(d *schema.ResourceData) *user.UserPostBody {
	body := new(user.UserPostBody)

	if username := d.Get("username"); username != nil {
		body.SetUsername(username.(string))
	}
	if firstname := d.Get("first_name"); firstname != nil {
		body.SetFirstName(firstname.(string))
	}
	if lastname := d.Get("last_name"); lastname != nil {
		body.SetLastName(lastname.(string))
	}
	if email := d.Get("email"); email != nil {
		body.SetEmail(email.(string))
	}
	if phone_number := d.Get("phone_number"); phone_number != nil {
		body.SetPhoneNumber(d.Get("phone_number").(string))
	}
	if password := d.Get("password"); password != nil {
		body.SetPassword(password.(string))
	}

	return body
}

func newUserPutBody(d *schema.ResourceData) *user.UserPutBody {
	body := new(user.UserPutBody)

	if username := d.Get("username"); username != nil {
		body.SetUsername(username.(string))
	}
	if firstname := d.Get("first_name"); firstname != nil {
		body.SetFirstName(firstname.(string))
	}
	if lastname := d.Get("last_name"); lastname != nil {
		body.SetLastName(lastname.(string))
	}
	if email := d.Get("email"); email != nil {
		body.SetEmail(email.(string))
	}
	if phone_number := d.Get("phone_number"); phone_number != nil {
		body.SetPhoneNumber(d.Get("phone_number").(string))
	}
	if password := d.Get("password"); password != nil {
		body.SetPassword(password.(string))
	}

	return body
}

func getUserWatchAttributes() []string {
	attributes := [...]string{
		"first_name", "last_name", "properties", "email", "phone_number",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getUserAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, user.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, user.ContextServerIndex, pco.server_index)
}

func decomposeUserId(d *schema.ResourceData, separator ...string) (string, string) {
	s := DecomposeResourceId(d.Id(), separator...)
	return s[0], s[1]
}
