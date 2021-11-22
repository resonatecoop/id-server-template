package config

import (
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/resonatecoop/user-api-client/client"
)

var (
	insecureSkipVerify = true // TODO keep this only for development
	basepath           = ""
	schemes            = []string{}
)

// NewAPIClient
func NewAPIClient(address, port string) *apiclient.ResonateServiceDocumentationUser {
	httpClient, err := httptransport.TLSClient(httptransport.TLSClientOptions{
		InsecureSkipVerify: insecureSkipVerify,
	})

	if err != nil {
		panic(err)
	}

	hostname := fmt.Sprintf("%s%s", address, port)
	transport := httptransport.NewWithClient(hostname, basepath, schemes, httpClient)

	client := apiclient.New(transport, strfmt.Default)

	return client
}
