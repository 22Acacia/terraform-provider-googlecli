package terraform-gcloud

import (
	"os"
	"bytes"
	"errors"
	"regexp"
	"strings"
	"os/exec"
	"io/ioutil"
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
func InitGcloud(config *Config) error {
	//  check that gcloud is installed
	_, err := exec.LookPath("gcloud")
	if err != nil {
		log.Println("gcloud cli is not installed.  Please install and try again")
		return err
	}

	//  ensure that the found gcloud is authorized
	account_file, err := setAccountFile(config.AccountFile)
	defer cleanupTempAccountFile(config.AccountFile, account_file)
	if err != nil {
		return err
	}
	auth_cmd := exec.Command("gcloud", "auth", "activate-service-account", "--key-file", account_file)
	var stdout, stderr bytes.Buffer
	auth_cmd.Stdout = &stdout
	auth_cmd.Stderr = &stderr
	err = auth_cmd.Run()
	if err != nil {
		log.Println("Dataflow auth failed with error: %q", stdout.String())
		return err 
	}
	
	// verify that datacloud functions are installed
	//  this will need to be updated when they come out of alpha
	datacloud_cmd := exec.Command("gcloud", "alpha", "dataflow" , "-h")
	err = datacloud_cmd.Run()
	if err != nil {
		log.Println("gcloud dataflow commands not installed.")
		return err
	}

	return nil
}

