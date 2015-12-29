package main

import (
	"os"
	"fmt"
	"time"
	"testing"
	"math/rand"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataflowCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataflowDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataflow,
				Check: resource.ComposeTestCheckFunc(
					testAccDataflowExists(
						"googlecli_dataflow.foobar"),
				),
			},
		},
	})
}

var disallowedDeletedStates = map[string]bool {
	"JOB_STATE_RUNNING": true,
	"JOB_STATE_UNKNOWN": true,
	"JOB_STATE_FAILED": true,
	"JOB_STATE_UPDATED": true,
}

func testAccCheckDataflowDestroy(s *terraform.State) error {
	projectName := os.Getenv("GOOGLE_PROJECT") 
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "googlecli_dataflow" {
			continue
		}

		jobdesc, err := ReadDataflow(rs.Primary.ID, projectName)
		if err != nil {
			return fmt.Errorf("Failed to read dataflow list")
		}

		if jobdesc.CurrentState == "" {
			return fmt.Errorf("Dataflow jobs never started ")
		}

		if _, ok := disallowedDeletedStates[jobdesc.CurrentState]; ok {
			return fmt.Errorf("Dataflow job in disallowed state: %q", jobdesc.CurrentState)
		}
	}

	return nil
}

var disallowedCreatedStates = map[string] bool {
	"JOB_STATE_FAILED": true,
	"JOB_STATE_STOPPED": true,
	"JOB_STATE_UPDATED": true,
	"JOB_STATE_UNKNOWN": true,
}

func testAccDataflowExists(n string) resource.TestCheckFunc {
	projectName := os.Getenv("GOOGLE_PROJECT") 
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		jobdesc, err := ReadDataflow(rs.Primary.Attributes["jobids.0"], projectName)
		if err != nil {
			return fmt.Errorf("In test: Command line read has errored: %q with rs.Primary hash: %q", err, rs.Primary)
		}

		if jobdesc.CurrentState == "" {
			return fmt.Errorf("Dataflow jobs never started")
		}

		if _, ok := disallowedCreatedStates[jobdesc.CurrentState]; ok {
			return fmt.Errorf("Dataflow job in disallowed state: %q", jobdesc.CurrentState)
		}

		return nil
	}
}

var randDataflowInt = rand.New(rand.NewSource(time.Now().UnixNano())).Int()

var testAccDataflow = fmt.Sprintf(`
resource "google_storage_bucket" "bucket" {
	name = "tf-test-bucket-%d"
	force_destroy = true
}
resource "googlecli_dataflow" "foobar" {
	name = "foobar-%d"
	classpath = "./test-fixtures/google-cloud-dataflow-java-examples-all-bundled-1.1.1-SNAPSHOT.jar"
	class = "com.google.cloud.dataflow.examples.WordCount"
	optional_args = {
		stagingLocation = "gs://${google_storage_bucket.bucket.name}"
		runner = "DataflowPipelineRunner"
	}
}
`, randDataflowInt, randDataflowInt)
