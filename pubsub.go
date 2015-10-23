package terraformGcloud

import (
	"fmt"
	"bytes"
	"strings"
	"os/exec"
	"github.com/hashicorp/terraform/helper/schema"
)

func CreatePubsub(d *schema.ResourceData) (string, error) {
	create_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "create", d.Get("name").(string), "--format", "json")
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

func ReadPubsub(d *schema.ResourceData) (bool, int, error) {
	read_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "list", "--format", "json")
	var stdout, stderr bytes.Buffer
	read_pubsub_cmd.Stdout = &stdout
	read_pubsub_cmd.Stderr = &stderr
	err := read_pubsub_cmd.Run()
	if err != nil {
		return false, 0, fmt.Errorf("Error listing pubsub topics: %q", stderr.String())
	}

	var pubsubList []map[string]string
	err = parseJSON(&pubsubList, stdout.String())
	if err != nil {
		return false, 0, err
	}

	pName, found := d.Get("name").(string), false
	for i := 0; i < len(pubsubList) && !found; i++ {
		if strings.Contains(pubsubList[i]["name"], pName) {
			found = true
		}
	}

	subcnt := 0
	if found == true {
		read_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptions", "list", "--format", "json")
		read_pubsub_cmd.Stdout = &stdout
		read_pubsub_cmd.Stderr = &stderr
		err := read_pubsub_cmd.Run()
		if err != nil {
			return false, 0, fmt.Errorf("Error listing pubsub subscriptions: %q", stderr.String())
		}
		
		var subscriptionList []map[string]string
		err = parseJSON(&subscriptionList, stdout.String())
		if err != nil {
			return found, 0, err
		}

		for _, doc := range subscriptionList {
			if strings.Contains(doc["topic"], pName) {
				subcnt++
			}
		}
	}

	return found, subcnt, nil
}

func DeletePubsub(d *schema.ResourceData) (error) {
	if d.Get("subscription_count").(int) > 1 {
		return fmt.Errorf("Topic has active subscriptions, will not delete")
	}

	delete_pubsub_cmd := exec.Command("gcloud", "alpha", "pubsub", "topics", "delete", d.Get("name").(string), "--format", "json")
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
