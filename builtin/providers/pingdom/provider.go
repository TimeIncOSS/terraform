package pingdom

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/russellcardullo/go-pingdom/pingdom"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PINGDOM_USERNAME", nil),
				Description: descriptions["username"],
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PINGDOM_PASSWORD", nil),
				Description: descriptions["password"],
			},

			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PINGDOM_API_KEY", nil),
				Description: descriptions["api_key"],
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"pingdom_http_check": resourcePingdomHttpCheck(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := pingdom.NewClient(
		d.Get("username").(string),
		d.Get("password").(string),
		d.Get("api_key").(string),
	)
	return client, nil
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"username": "",
		"password": "",
		"api_key":  "",
	}
}
