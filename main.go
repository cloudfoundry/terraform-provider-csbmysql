package main

import (
	"github.com/cloudfoundry/terraform-provider-csbmysql/csbmysql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbmysql.Provider,
	})
}
