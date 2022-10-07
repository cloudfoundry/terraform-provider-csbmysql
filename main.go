package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/cloudfoundry/terraform-provider-csbmysql/csbmysql"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbmysql.Provider,
	})
}
