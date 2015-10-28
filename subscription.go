package terraformGcloud

import (
	"fmt"
	"bytes"
	"strings"
	"os/exec"
)

func CreateSubscription(name, topic string) (string, error) {
	create_subscription_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptions", "create", name, "--topic", topic, "--format", "json")
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

func ReadSubscription(name string) (string, error) {
	read_subscription_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptions", "list", "--format", "json")
	var stdout, stderr bytes.Buffer
	read_subscription_cmd.Stdout = &stdout
	read_subscription_cmd.Stderr = &stderr
	err := read_subscription_cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error listing pubsub subscriptions: %q", stderr.String())
	}

	var subscriptionList []subscriptionElem
	err = parseJSON(&subscriptionList, stdout.String())
	if err != nil {
		return "", err
	}

	nameArr := strings.Split(name, "/")
	sName, fullname := nameArr[len(nameArr)-1], ""
	for i := 0; i < len(subscriptionList) && fullname == ""; i++ {
		if strings.Contains(subscriptionList[i].Name, sName) {
			fullname = subscriptionList[i].Name
		}
	}

	return fullname, nil
}

func DeleteSubscription(name string) (error) {
	delete_subscription_cmd := exec.Command("gcloud", "alpha", "pubsub", "subscriptions", "delete", name, "--format", "json")
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
