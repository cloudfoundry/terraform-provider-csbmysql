package csbmysql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema:               ProviderSchema(),
		ConfigureContextFunc: ProviderConfigureContext,
		ResourcesMap: map[string]*schema.Resource{
			ResourceNameKey: ResourceBindingUser(),
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
		sslRootCertKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		sslCertKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		sslKeyKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		skipVerifyKey: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "skip_verify controls whether a client verifies the server's certificate chain and host name. If skip_verify is true, crypto/tls accepts any certificate presented by the server and any host name in that certificate.",
		},
	}
}

func ProviderConfigureContext(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	factory := connectionFactory{
		host:                        d.Get(hostKey).(string),
		port:                        d.Get(portKey).(int),
		username:                    d.Get(usernameKey).(string),
		password:                    d.Get(passwordKey).(string),
		database:                    d.Get(databaseKey).(string),
		caCertificate:               []byte(d.Get(sslRootCertKey).(string)),
		clientCertificate:           []byte(d.Get(sslCertKey).(string)),
		clientCertificatePrivateKey: []byte(d.Get(sslKeyKey).(string)),
		skipVerify:                  d.Get(skipVerifyKey).(bool),
	}

	return factory, diags
}
