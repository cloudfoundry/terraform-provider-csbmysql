package csbmysql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	databaseKey  = "database"
	passwordKey  = "password"
	usernameKey  = "username"
	portKey      = "port"
	hostKey      = "host"
	tlsKey       = "require_ssl"
	ResourceName = "csbmysql_binding_user"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema:               ProviderSchema(),
		ConfigureContextFunc: ProviderConfigureContext,
		ResourcesMap: map[string]*schema.Resource{
			ResourceName: ResourceBindingUser(),
		},
	}
}

func ProviderSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		hostKey: {
			Type:     schema.TypeString,
			Required: true,
		},
		portKey: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IsPortNumber,
		},
		usernameKey: {
			Type:     schema.TypeString,
			Required: true,
		},
		passwordKey: {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		databaseKey: {
			Type:     schema.TypeString,
			Required: true,
		},
		tlsKey: {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}

func ProviderConfigureContext(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	factory := connectionFactory{
		host:      d.Get(hostKey).(string),
		port:      d.Get(portKey).(int),
		username:  d.Get(usernameKey).(string),
		password:  d.Get(passwordKey).(string),
		database:  d.Get(databaseKey).(string),
		verifyTLS: d.Get(tlsKey).(bool),
	}

	return factory, diags
}
