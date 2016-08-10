package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSesDomainIdentity() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesDomainIdentityCreate,
		Read:   resourceAwsSesDomainIdentityRead,
		Delete: resourceAwsSesDomainIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"verify_dkim": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"dkim_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"verification_token": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"dkim_tokens": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
			},
		},
	}
}

func resourceAwsSesDomainIdentityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	domainInput := ses.VerifyDomainIdentityInput{
		Domain: aws.String(""),
	}
	domainOut, err := conn.VerifyDomainIdentity(&domainInput)
	domainOut.VerificationToken

	dkimInput := ses.VerifyDomainDkimInput{
		Domain: aws.String(""),
	}
	dkimOut, err := conn.VerifyDomainDkim(dkimInput)
	dkimOut.DkimTokens

	input := ses.SetIdentityMailFromDomainInput{}
	input.Identity
	input.BehaviorOnMXFailure
	input.MailFromDomain
	conn.SetIdentityMailFromDomain(input)

	return resourceAwsSesDomainIdentityUpdate(d, meta)
}

func resourceAwsSesDomainIdentityUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn
	input := ses.SetIdentityDkimEnabledInput{}
	input.Identity
	input.DkimEnabled
	conn.SetIdentityDkimEnabled(input)
}

func resourceAwsSesDomainIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	inputIdentities := ses.ListIdentitiesInput{}
	inputIdentities.IdentityType
	conn.ListIdentities(&inputIdentities)

	input := ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{d.Id()},
	}
	out, err := conn.GetIdentityVerificationAttributes(&input)
	out.VerificationAttributes[""].VerificationStatus
	out.VerificationAttributes[""].VerificationToken
}

func resourceAwsSesDomainIdentityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	input := ses.DeleteIdentityInput{
		Identity: aws.String(""),
	}
	_, err := conn.DeleteIdentity(&input)
	if err != nil {
		return err
	}

	return nil
}
