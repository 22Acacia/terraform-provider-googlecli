package terraformGcloud

import (
	"fmt"
	"bytes"
	"strings"
	"os/exec"
)

func CreatePubsub(pubsubName string) (string, error) {
	create_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "create", pubsubName, "--format", "json")
	var stdout, stderr bytes.Buffer
	create_pubsub_cmd.Stdout = &stdout
	create_pubsub_cmd.Stderr = &stderr
	err := create_pubsub_cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error creating pubsub: %q", stderr.String())
	}

	var pubsubRet [][]interface{}
	err = parseJSON(&pubsubRet, stdout.String())
	if err != nil {
		return "", fmt.Errorf("Failed to deserialize %q with error: %q", stdout.String(), err)
	}

	if len(pubsubRet[1]) > 0 {
		return "", fmt.Errorf("Error creating pubsub: %q", pubsubRet[1])
	} 
	
	success := pubsubRet[0][0].(map[string]interface{})
	return success["name"].(string), nil
}

func ReadPubsub(pubsubName string) (string, int, error) {
	read_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "list", "--format", "json")
	var stdout, stderr bytes.Buffer
	read_pubsub_cmd.Stdout = &stdout
	read_pubsub_cmd.Stderr = &stderr
	err := read_pubsub_cmd.Run()
	if err != nil {
		return "", 0, fmt.Errorf("Error listing pubsub topics: %q", stderr.String())
	}

	var pubsubList []map[string]string
	err = parseJSON(&pubsubList, stdout.String())
	if err != nil {
		return "", 0, err
	}
	
	pubsubArr := strings.Split(pubsubName, "/")
	pName, fullname := pubsubArr[len(pubsubArr)-1], ""
	for i := 0; i < len(pubsubList) && fullname == ""; i++ {
		if strings.Contains(pubsubList[i]["name"], pName) {
			fullname = pubsubList[i]["name"]
		}
	}

	subcnt := 0
	if fullname != "" {
		read_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "list-subscriptions", pName, "--format", "json")
		stdout.Reset()
		stderr.Reset()
		read_pubsub_cmd.Stdout = &stdout
		read_pubsub_cmd.Stderr = &stderr
		err := read_pubsub_cmd.Run()
		if err != nil {
			return fullname, 0, fmt.Errorf("Error listing pubsub subscriptions: %q", stderr.String())
		}
		
		var subscriptionList []string
		err = parseJSON(&subscriptionList, stdout.String())
		if err != nil {
			return fullname, 0, fmt.Errorf("failed string: %q with error: %q", stdout.String(), err)
		}

		subcnt = len(subscriptionList)
	}

	return fullname, subcnt, nil
}

func DeletePubsub(pubsubName string, subCount int) (error) {
	if subCount > 0 {
		return fmt.Errorf("Topic has active subscriptions, will not delete")
	}

	delete_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "delete", pubsubName, "--format", "json")
	var stdout, stderr bytes.Buffer
	delete_pubsub_cmd.Stdout = &stdout
	delete_pubsub_cmd.Stderr = &stderr
	err := delete_pubsub_cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to delete pubsub topic: %q", stderr.String())
	}

	var pubsubRet [][]interface{}
	err = parseJSON(&pubsubRet, stdout.String())
	if err != nil {
		return err
	}

	if len(pubsubRet[1]) > 0 {
		return fmt.Errorf("Error deleting pubsub: %q", pubsubRet)
	}

	return nil
}
