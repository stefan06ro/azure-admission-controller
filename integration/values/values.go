package values

import (
	"bytes"
	"text/template"

	"github.com/giantswarm/microerror"
)

const Template = `Installation:
  V1:
    Guest:
      Kubernetes:
        API:
          EndpointBase: k8s.test.westeurope.azure.gigantic.io
    Provider:
      Azure:
        Location: westeurope
    Registry:
      Domain: quay.io
    Secret:
      Credentiald:
        Azure:
          CredentialDefault:
            ClientID: {{.ClientID}}
            ClientSecret: {{.ClientSecret}}
            TenantID: {{.TenantID}}
            SubscriptionID: {{.SubscriptionID}}
`

type Credentials struct {
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
}

func YAML(clientID, clientSecret, subscriptionID, tenantID string) (string, error) {
	tmpl, err := template.New("values").Parse(Template)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, Credentials{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	return tpl.String(), nil
}
