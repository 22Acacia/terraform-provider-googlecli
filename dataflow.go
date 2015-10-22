package terraformGcloud

import (
	"fmt"
	"bytes"
	"regexp"
	"strings"
	"os/exec"
	"github.com/hashicorp/terraform/helper/schema"
)

type dataflowDescription struct {
	ClientRequestId		string	`json:"clientRequestId"`
	CreateTime		string	`json:"createTime"`
	CurrentState		string	`json:"currentState"`
	CurrentStateTime	string	`json:"currentStateTime"`
	Id			string	`json:"id"`
	Name			string	`json:"name"`
	ProjectId		string	`json:"projectId"`
	Type			string	`json:"type"`
}


func CreateDataflow(d *schema.ResourceData) ([]string, error) {
	//  at this point we have verified that our command line jankiness is going to work
	//  get to it
	//  I'm assuming, possibly foolishly, that java is installed on this system
	create_dataflow_cmd := exec.Command("java", "-cp", d.Get("jarfile").(string), d.Get("class").(string), "--project="+d.Get("project").(string), "--stagingLocation="+d.Get("staging_bucket").(string), "--jobName="+d.Get("name").(string))
	var stdout, stderr bytes.Buffer
	create_dataflow_cmd.Stdout = &stdout
	create_dataflow_cmd.Stderr = &stderr
	err := create_dataflow_cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error submitting dataflow job: %q", stderr.String())
	}

	//  job successfully submitted, now get the job id
	jobidRe := regexp.MustCompile("Submitted job: ([0-9-_]+)\n")
	jobidmatches := jobidRe.FindAllStringSubmatch(stdout.String(), -1)
	jobids := make([]string, 0)
	for _, match := range jobidmatches {
		jobids = append(jobids, match[1])
	}

	return jobids, nil
}

func ReadDataflow(d *schema.ResourceData) ([]string, error) {
	job_states := make([]string, 0)
	for i := 0; i < d.Get("jobids.#").(int); i++ {
		key := fmt.Sprintf("jobids.%d", i)
		job_check_cmd := exec.Command("gcloud", "alpha", "dataflow", "jobs", "describe", d.Get(key).(string), "--format", "json")
		var stdout, stderr bytes.Buffer
		job_check_cmd.Stdout = &stdout
		job_check_cmd.Stderr = &stderr
		err := job_check_cmd.Run()
		if err != nil {
			return nil, err
		}

		var jobDesc dataflowDescription
		fmt.Println(stdout.String())
		err = parseJSON(&jobDesc, stdout.String())
		if err != nil {
			return nil, err
		}
		job_states = append(job_states, jobDesc.CurrentState)
	}

	return job_states, nil
}

func CancelDataflow(d *schema.ResourceData) ([]string, error) {
	failedCancel := make([]string, 0)
	for i := 0; i < d.Get("jobids.#").(int); i++ {
		jobstatekey := fmt.Sprintf("job_states.%d", i)
		jobstate := d.Get(jobstatekey).(string)
		if jobstate == "JOB_STATE_RUNNING" {
			jobidkey := fmt.Sprintf("jobids.%d", i)
			job_cancel_cmd := exec.Command("gcloud", "alpha", "dataflow", "jobs", "cancel", d.Get(jobidkey).(string))
			var stdout, stderr bytes.Buffer
			job_cancel_cmd.Stdout = &stdout
			job_cancel_cmd.Stderr = &stderr
			err := job_cancel_cmd.Run()
			if err != nil {
				return nil, err
			}

			if strings.Contains(stdout.String(), "Failed") {
				failedCancel = append(failedCancel,d.Get(jobidkey).(string))
			}
		}
	}

	return failedCancel, nil
}
