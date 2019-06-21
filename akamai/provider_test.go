package akamai

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// AKAMAI_ACCESS_TOKEN
// AKAMAI_CLIENT_SECRET
// AKAMAI_CLIENT_TOKEN
// AKAMAI_HOST
var credsEnvVars = []string{
	"AKAMAI_ACCESS_TOKEN",
	"AKAMAI_CLIENT_SECRET",
	"AKAMAI_CLIENT_TOKEN",
	"AKAMAI_HOST",
}

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"akamai": testAccProvider,
	}

}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if envf := os.Getenv("AKAMAI_ENVRC_FILE"); envf != "" {
		_, err := ioutil.ReadFile(envf)
		if err != nil {
			t.Fatalf("Error reading AKAMAI_ENVRC_FILE path: %s", err)
		}
	}

	// Make sure we can read from env variables if we haven't specified AKAMAI_ENVRC_FILE
	for _, k := range credsEnvVars {
		if v := os.Getenv(k); v == "" {
			t.Fatalf("%s must be set if .envrc file not found.", k)
		}
	}
}

/*
func testAccAkamaiProviderHost(provider *schema.Provider) string {
	if provider == nil {
		log.Print("[DEBUG] unable to read Akamai Host from test provider: empty provider")
		return ""
	}

	if provider.Meta() == nil {
		log.Print("[DEBUG] unable to read Akamai Host from test provider: unconfigured provider")
		return ""
	}

	ac, ok := provider.Meta().(*AkamaiClient)
	if !ok {
		log.Print("[DEBUG] Unable to read Akamai Host from test provider: non-Akamai or unconfigured Akamai provider")
		return ""
	}

	creds, err := ac.client.Credentials.Get()
	if err != nil {
		log.Print("[DEBUG] Unable to read Akamai Host from test provider: could not access Akamai Credentials.")
		return ""
	}

	return creds.Host
}

// testAccGetAkamaiHost returns the akamai host of the testAccProvider
// Must be used returned within a resource.TestCheckFunc
func testAccGetAkamaiHost() string {
	return testAccAkamaiProviderHost(testAccProvider)
}

*/
