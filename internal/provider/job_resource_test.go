package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJobResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobResourceConfig("Demo Form Ansible No input"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "form_name", "Demo Form Ansible No input"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.name", "github.com/dsha256"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.opco", "myopco"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.svm_name", "mysvm_name"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.exposure", "myexposure"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.env", "myenv"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.dataclass", "mydataclass"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.share_name", "myshare_name"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.accountid", "myaccountid"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.size", "mysize"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.protection_required", "myprotection_required"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "credentials.ontap_cred", "myontap_cred"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "credentials.bind_cred", "mybind_cred"),
					// Check that an ID has been set (we don't know what the value is as it changes
					resource.TestCheckResourceAttrSet("ansible-forms_job_resource.job", "id"),
					resource.TestCheckResourceAttr("ansible-forms_job_resource.job", "extravars.region", "myregion")),
			},
			{
				Config:      testAccJobResourceConfig("Non Existent Form Name"),
				ExpectError: regexp.MustCompile("Error running apply"),
			},
		},
	})
}

func testAccJobResourceConfig(jobFormName string) string {
	host := os.Getenv("TF_ACC_ANSIBLE_FORMS_HOST")
	//host := "127.0.0.1:8443"
	admin := os.Getenv("TF_ACC_ANSIBLE_FORMS_USER")
	//admin := "admin"
	password := os.Getenv("TF_ACC_ANSIBLE_FORMS_PASS")
	//password := "AnsibleForms!123"
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_ANSIBLE_FORMS_HOST, TF_ACC_ANSIBLE_FORMS_USER, and TF_ACC_ANSIBLE_FORMS_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "ansible-forms" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "ansible-forms_job_resource" "job" {
 cx_profile_name = "cluster4"
  form_name       = "%s"
  extravars = {
    name                = "github.com/dsha256"
    region              = "myregion"
    opco                = "myopco"
    svm_name            = "mysvm_name"
    exposure            = "myexposure"
    env                 = "myenv"
    dataclass           = "mydataclass"
    share_name          = "myshare_name"
    accountid           = "myaccountid"
    size                = "mysize"
    protection_required = "myprotection_required"
  }
  credentials = {
    ontap_cred = "myontap_cred"
    bind_cred  = "mybind_cred"
  }
}`, host, admin, password, jobFormName)
}
