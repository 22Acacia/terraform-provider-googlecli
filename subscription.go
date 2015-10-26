package terraformGcloud

import (
	"fmt"
	"bytes"
	"strings"
	"os/exec"
	"github.com/hashicorp/terraform/helper/schema"
)

func CreateSubscription(d *schema.ResourceData) (string, error) {
	create_subscription_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptions", "create", d.Get("name").(string), "--topic", d.Get("topic").(string), "--format", "json")
	var stdout, stderr bytes.Buffer
	create_subscription_cmd.Stdout = &stdout
	create_subscription_cmd.Stderr = &stderr
	err := create_subscription_cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error creating subscription: %q", stderr.String())
	}

	var subscriptionRet [][]interface{}
	err = parseJSON(&subscriptionRet, stdout.String())
	if err != nil {
		return "", fmt.Errorf("Failed to deserialize %q with error: %q", stdout.String(), err)
	}

	if len(subscriptionRet[1]) > 0 {
		return "", fmt.Errorf("Error creating subscription: %q", subscriptionRet[1])
	} 
	
	success := subscriptionRet[0][0].(map[string]interface{})
	return success["name"].(string), nil
}

type subscriptionElem struct {
	AckDeadlineSeconds	int		`json:"ackDeadlineSeconds"`
	Name			string		`json:"name"`
	PushConfig		interface{}	`json:"pushConfig"`
	Topic			string		`json:"topic"`
}

func ReadSubscription(d *schema.ResourceData) (bool, error) {
	read_subscription_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptions", "list", "--format", "json")
	var stdout, stderr bytes.Buffer
	read_subscription_cmd.Stdout = &stdout
	read_subscription_cmd.Stderr = &stderr
	err := read_subscription_cmd.Run()
	if err != nil {
		return false, fmt.Errorf("Error listing pubsub subscriptions: %q", stderr.String())
	}

	var subscriptionList []subscriptionElem
	err = parseJSON(&subscriptionList, stdout.String())
	if err != nil {
		return false, err
	}

	sName, found := d.Get("name").(string), false
	for i := 0; i < len(subscriptionList) && !found; i++ {
		if strings.Contains(subscriptionList[i].Name, sName) {
			found = true
		}
	}

	return found, nil
}

func DeleteSubscription(d *schema.ResourceData) (error) {
	delete_subscription_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptionss", "delete", d.Get("name").(string), "--format", "json")
	var stdout, stderr bytes.Buffer
	delete_subscription_cmd.Stdout = &stdout
	delete_subscription_cmd.Stderr = &stderr
	err := delete_subscription_cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to delete pubsub subscription: %q", stderr.String())
	}

	var subscriptionRet [][]interface{}
	err = parseJSON(&subscriptionRet, stdout.String())
	if err != nil {
		return err
	}

	if len(subscriptionRet[1]) > 0 {
		return fmt.Errorf("Error deleting subscription: %q", subscriptionRet)
	}

	return nil
}
