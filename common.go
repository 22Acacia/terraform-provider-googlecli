package terraformGcloud

import (
	"os"
	"fmt"
	"bytes"
	"strings"
	"os/exec"
	"io/ioutil"
	"encoding/json"
)

// accountFile represents the structure of the account file JSON file.
type accountFile struct {
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
}

func parseJSON(result interface{}, contents string) error {
	r := strings.NewReader(contents)
	dec := json.NewDecoder(r)

	return dec.Decode(result)
}

//  return pointer to a file that contains account information
//   this was swiped directly from terraform, its works.  its fine
func setAccountFile(contents string) (string, error) {
	if contents != "" {
		var account accountFile
		if err := parseJSON(&account, contents); err == nil {
			//  raw account info, write out to a file
			tmpfile, err := ioutil.TempFile("","")
			if err != nil {
				return "", err
			}
			_, err = tmpfile.WriteString(contents)
			if err != nil {
				return "", err
			}
			tmpfile.Close()
			return tmpfile.Name(), nil
		} else {
			return contents, nil
		}
	}
	return "", nil
}

func cleanupTempAccountFile(rawAccountFile, account_file string) {
	if rawAccountFile != account_file {
		os.Remove(account_file)
	}
}

//  init function will make sure that gcloud cli is installed,
//  authorized and that dataflow commands are available

func InitGcloud(accountFileRaw string) error {
	//  check that gcloud is installed
	_, err := exec.LookPath("gcloud")
	if err != nil {
		return fmt.Errorf("gcloud cli is not installed.  Please install and try again\n")
	}

	//  check that java is installed
	_, err = exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("java jre (at least) is not installed.  Please install and try again\n")
	}

	//  ensure that the found gcloud is authorized
	account_file, err := setAccountFile(accountFileRaw)
	defer cleanupTempAccountFile(accountFileRaw, account_file)
	if err != nil {
		return err
	}
	auth_cmd := exec.Command("gcloud", "auth", "activate-service-account", "--key-file", account_file)
	var stdout, stderr bytes.Buffer
	auth_cmd.Stdout = &stdout
	auth_cmd.Stderr = &stderr
	err = auth_cmd.Run()
	if err != nil {
		return fmt.Errorf("gcloud auth failed with error: %s\n", stderr.String())
	}
	
	// verify that datacloud functions are installed
	//  this will need to be updated when they come out of alpha
	datacloud_cmd := exec.Command("gcloud", "alpha", "dataflow" , "-h")
	err = datacloud_cmd.Run()
	if err != nil {
		return fmt.Errorf("gcloud dataflow commands not installed.\n")
	}

	return nil
}

//  kubectl is only used when working with pods in a container so we'll check it on its own
func InitKubectl(container, project, zone string) error {
	//  check that kubectl is installed
	_, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl is not installed.  Please install and try again\n")
	}

	cred_gen_cmd := exec.Command("gcloud", "beta",  "container", "clusters", "get-credentials", container, "--project=" + project, "--zone=" + zone)
	var stdout, stderr bytes.Buffer
	cred_gen_cmd.Stdout = &stdout
	cred_gen_cmd.Stderr = &stderr
	err = cred_gen_cmd.Run()
	if err != nil {
		return fmt.Errorf("Gcloud container credential fetch failed: %s\n", stderr.String())
	}

	
	kubectl_check_cmd := exec.Command("kubectl", "config", "view")
	kubectl_check_cmd.Stdout = &stdout
	kubectl_check_cmd.Stderr = &stderr
	err = kubectl_check_cmd.Run()
	if err != nil {
		return fmt.Errorf("Kubectl config view command failed: %q\n", stderr.String())
	}
	
	return nil
}
