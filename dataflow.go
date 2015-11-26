package terraformGcloud

import (
	"fmt"
	"time"
	"bytes"
	"regexp"
	"strings"
	"os/exec"
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


func CreateDataflow(name, classpath, class, project string, optional_args map[string]string) ([]string, error) {
	//  at this point we have verified that our command line jankiness is going to work
	//  get to it
	dataflow_cmd := "java"
	dataflow_args := []string{"-cp", classpath, class, "--jobName=" + name, "--project=" + project}
	for k, v := range optional_args {
		dataflow_args = append(dataflow_args, "--" + k + "=" + v)
	}

	create_dataflow_cmd := exec.Command(dataflow_cmd, dataflow_args...)
	var stdout, stderr bytes.Buffer
	create_dataflow_cmd.Stdout = &stdout
	create_dataflow_cmd.Stderr = &stderr
	err := create_dataflow_cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error submitting dataflow job: %q", stderr.String())
	}

	return findJobIds(stdout.String()), nil
}

func findJobIds(creation_stdout string) ([]string) {
	//  job successfully submitted, now get the job id
	jobidRe := regexp.MustCompile("Submitted job: ([0-9-_]+).*")
	jobidmatches := jobidRe.FindAllStringSubmatch(creation_stdout, -1)
	jobids := make([]string, 0)
	for _, match := range jobidmatches {
		jobids = append(jobids, match[1])
	}

	return jobids
}

func ReadDataflow(jobkey string) (string, error) {
	//  we will often read the job as we create it, but the state doesn't get set immediately so we
	//  end up saving "" as the state.  which is bad times.  sleep three seconds to wait for status
	//  to be set
	time.Sleep(3 * time.Second)
	job_check_cmd := exec.Command("gcloud", "alpha", "dataflow", "jobs", "describe", jobkey, "--format", "json")
	var stdout, stderr bytes.Buffer
	job_check_cmd.Stdout = &stdout
	job_check_cmd.Stderr = &stderr
	err := job_check_cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error reading job %q with error %q", jobkey, stderr.String())
	}

	var jobDesc dataflowDescription
	err = parseJSON(&jobDesc, stdout.String())
	if err != nil {
		return "", err
	}
	job_state := jobDesc.CurrentState

	return job_state, nil
}

func CancelDataflow(jobid, jobstate string) (bool, error) {
	failedCancel := false
	if jobstate == "JOB_STATE_RUNNING" {
		job_cancel_cmd := exec.Command("gcloud", "alpha", "dataflow", "jobs", "cancel", jobid)
		var stdout, stderr bytes.Buffer
		job_cancel_cmd.Stdout = &stdout
		job_cancel_cmd.Stderr = &stderr
		err := job_cancel_cmd.Run()
		if err != nil {
			return false, err
		}

		if strings.Contains(stdout.String(), "Failed") {
			failedCancel = true
		}
	}

	return failedCancel, nil
}
