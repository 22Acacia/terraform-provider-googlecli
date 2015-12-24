package main

import (
	"fmt"
	"time"
	"bytes"
	"regexp"
	"strings"
	"os/exec"
)

type DataflowDescription struct {
	CurrentState	string	`json:"currentState"`
	RequestedState	string	`json:"requestedState"`
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
		return findJobIds(stderr.String()), fmt.Errorf("Error submitting dataflow job: %q", stderr.String())
	}

	return findJobIds(stdout.String()), nil
}

func findJobIds(creation_stdout string) ([]string) {
	//  job successfully submitted, now get the job id
	jobidRe := regexp.MustCompile("(\\d{4}-\\d{2}-\\d{2}_\\d{2}_\\d{2}_\\d{2}-\\d{10,})")
	jobidmatches := jobidRe.FindAllStringSubmatch(creation_stdout, -1)
	jobids := make([]string, 0)
	for _, match := range jobidmatches {
		jobids = append(jobids, match[1])
	}

	return jobids
}

func ReadDataflow(jobkey string, project string) (*DataflowDescription, error) {
	//  we will often read the job as we create it, but the state doesn't get set immediately so we
	//  end up saving "" as the state.  which is bad times.  sleep five seconds to wait for status
	//  to be set
	time.Sleep(5 * time.Second)
	job_check_cmd := exec.Command("gcloud", "alpha", "dataflow", "jobs", "describe", jobkey, "--format=json", "--project=" +project)
	var stdout, stderr bytes.Buffer
	job_check_cmd.Stdout = &stdout
	job_check_cmd.Stderr = &stderr
	err := job_check_cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Error reading job %q with error %q", jobkey, stderr.String())
	}

	jobDesc := &DataflowDescription{}
	err = parseJSON(jobDesc, stdout.String())
	if err != nil {
		return nil, err
	}

	return jobDesc, nil
}

func CancelDataflow(jobid, jobstate string, project string) (bool, error) {
	failedCancel := false
	if jobstate == "JOB_STATE_RUNNING" {
		job_cancel_cmd := exec.Command("gcloud", "alpha", "dataflow", "jobs", "cancel", jobid, "--project="+project)
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
