package main

import (
	"io"
	"os"
	"fmt"
	"time"
	"strings"
	"crypto/md5"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDataflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataflowCreate,
		Read:   resourceDataflowRead,
		Delete: resourceDataflowDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"classpath": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"class": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"optional_args": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:	  schema.TypeString,
			},

			"jobids": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"job_states": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			
			"jar_checksums": &schema.Schema{
				Type:	  schema.TypeMap,
				Computed: true,
				ForceNew: true,
				Elem: 	  schema.TypeString,
			},
		},
	}
}

func resourceDataflowCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	optional_args := cleanAdditionalArgs(d.Get("optional_args").(map[string]interface{}))

	jobids, err := CreateDataflow(d.Get("name").(string), d.Get("classpath").(string), d.Get("class").(string), config.Project, optional_args)
	if err != nil && len(jobids) == 0 {
		// call failed, abort
		return err
	} else if err != nil {
		// we're updating, check and make sure all jobs found have been cancelled, if not, quit
		for _, jobid := range jobids {
			jobdesc, err := ReadDataflow(jobid, config.Project)
			if err != nil {
				return err
			}
			if jobdesc.RequestedState == "JOB_STATE_RUNNING" || jobdesc.RequestedState == "JOB_STATE_UPDATED" {
				return fmt.Errorf("Attempting to create existing job name %s but prior job of same name with id %s still exists and is in state %s which we won't blindly overwrite.", d.Get("name").(string),  jobid, jobdesc.CurrentState)
			}
		}

		// wait for 10 minutes or all jobs cancelled/dead/unknown
		not_all_toast := true
		for i := 0; i < (10 * 6) && not_all_toast; i++ {
			time.Sleep(10 * time.Second)
			not_all_toast = false
			//  check all jobs, if not in a cancelled state, set state flag
			for _, jobid := range jobids {
				jobdesc, err := ReadDataflow(jobid, config.Project)
				if err != nil {
					return err
				}
				if jobdesc.RequestedState == "JOB_STATE_RUNNING" || jobdesc.RequestedState == "JOB_STATE_UPDATED" {
					not_all_toast = true
				}
			}
		}

		if not_all_toast {
			return fmt.Errorf("Not all jobs entered into a cancelled/dead/unknown state but all jobs have been requested to be cancelled.  Please wait a few minutes and try again.")
		}

		//  retry the job creation, any errors here and abort
		//  dataflow won't let us reuse job names so at this point gen a timestamp as a string and append
		//    to the supplied name and go with god.  we use jobids for all the work anyway.
		nameTimestamp := time.Now().Format("2006-01-02T15-04-05Z07-00")
		jobids, err = CreateDataflow(d.Get("name").(string) + "." + nameTimestamp, d.Get("classpath").(string), d.Get("class").(string), config.Project, optional_args)
		if err != nil {
			return fmt.Errorf("Failed to createa job named %q with error: %q", nameTimestamp,err)
		}
	}
	

	d.Set("jobids", jobids)
	d.SetId(d.Get("name").(string))

	err = resourceDataflowRead(d, meta)
	if err != nil {
		return err
	}

	return nil
}

func resourceDataflowRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	//  determine job states
	job_states := make([]string, 0)
	job_cnt := d.Get("jobids.#")
	if job_cnt != nil {
		for i := 0; i < job_cnt.(int); i++ {
			jobidkey:= fmt.Sprintf("jobids.%d", i)
			job_desc, err := ReadDataflow(d.Get(jobidkey).(string), config.Project)
			if err != nil {
				return err
			}
			job_states = append(job_states, job_desc.CurrentState)
		}
	}

	//  calculate md5 of each jar in classpath so we can regen dataflow job on change
	jars := strings.Splits(d.Get("classpath"), ";")
	jar_checksums := []map[string]string
	for _, jar := range jars {
		jar_checksums[jar] := computeMd5(jar)
	}

	d.Set("job_states", job_states)
	d.Set("jar_checksums", jar_checksums)

	return nil
}

//  function swiped from:  http://dev.pawelsz.eu/2014/11/google-golang-compute-md5-of-file.html
func computeMd5(filePath string) (string, error) {
 	file, err := os.Open(filePath)
 	if err != nil {
    	return "", err
 	}
  	defer file.Close()
 
  	hash := md5.New()
  	if _, err := io.Copy(hash, file); err != nil {
    	return "", err
  	}
 
  	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func resourceDataflowDelete(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*Config)
	err := resourceDataflowRead(d, meta)
	if err != nil {
		return err
	}

	failedCancel := make([]string, 0)
	job_cnt := d.Get("jobids.#")
	if job_cnt != nil {
		for i := 0; i < job_cnt.(int); i++ {
			jobidkey:= fmt.Sprintf("jobids.%d", i)
			jobstatekey := fmt.Sprintf("job_states.%d", i)
			failedjob, err := CancelDataflow(d.Get(jobidkey).(string), d.Get(jobstatekey).(string), config.Project)
			if err != nil {
				return err
			}
			if failedjob {
				failedCancel = append(failedCancel, d.Get(jobidkey).(string))
			}
		}
	}

	if len(failedCancel) > 0 {
		return fmt.Errorf("Failed to cancel the following jobs: %v", failedCancel)
	}

	d.SetId("")
	return nil
}
