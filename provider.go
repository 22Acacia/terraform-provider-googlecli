package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.EnvDefaultFunc("GOOGLE_CREDENTIALS", nil),
			},

			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GOOGLE_PROJECT", nil),
			},

			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GOOGLE_REGION", nil),
			},

			"CredentialsFile": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"googlecli_dataflow":                       resourceDataflow(),
			"googlecli_container_replica_controller":   resourceContainerReplicaController(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Credentials: d.Get("credentials").(string),
		Project:     d.Get("project").(string),
		Region:      d.Get("region").(string),
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, err
	}
	defer config.cleanupTempAccountFile()

	//  init gcloud cli
	if err := config.initGcloud(); err != nil {
		return nil, err
	}

	return &config, nil
}

func cleanAdditionalArgs(optional_args map[string]interface{}) map[string]string {
	cleaned_opts := make(map[string]string)
	for k,v := range  optional_args {
		cleaned_opts[k] = v.(string)
	}
	return cleaned_opts
}

