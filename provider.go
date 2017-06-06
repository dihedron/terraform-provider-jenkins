package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider creates a new JenkinsCI provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JENKINS_URL", nil),
				Description: "The URL of the JenkinsCI server to connect to.",
			},
			"ca_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JENKINS_CA_CERT", nil),
				Description: "The path to the JenkinsCi self-signed certificate.",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JENKINS_USERNAME", nil),
				Description: "Username to authenticate to JenkinsCI.",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JENKINS_PASSWORD", nil),
				Description: "Password to authenticate to JenkinsCI.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"jenkins_job": resourceJenkinsJob(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		ServerURL: d.Get("server_url").(string),
		CACert:    d.Get("ca_cert").(string),
		Username:  d.Get("username").(string),
		Password:  d.Get("password").(string),
	}

	client, err := config.getAPIClient()
	if err != nil {
		return nil, err
	}

	return client, nil
}
