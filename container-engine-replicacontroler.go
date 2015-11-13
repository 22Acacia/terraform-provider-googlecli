package terraformGcloud

import (
	"fmt"
	"bytes"
	"os/exec"
)


type kubectlItem struct {
	Metadata	struct {
		Name 		string	`json:"name"`
		Uid		string	`json:"uid"`
	} 				`json:"metadata"`
	Spec	interface{}	`json:"spec"`
	Status	 struct {
		Replicas	int	`json:"replicas"`
	}				`json:"status"`
}


func CreateKubeRC(name, dockerImage string, optional_args map[string]string) (string, error) {
	kubectl_cmd := "kubectl"
	kubectl_run_args :=[]string{"run", name, "--image=" + dockerImage, "--output=json"}
	for k, v := range optional_args {
		kubectl_run_args = append(kubectl_run_args, "--" + k + "=" +v)
	}
	run_replicacontroler := exec.Command(kubectl_cmd, kubectl_run_args...)
	var stdout, stderr bytes.Buffer
	run_replicacontroler.Stdout = &stdout
	run_replicacontroler.Stderr = &stderr
	err := run_replicacontroler.Run()
	if err != nil {
		return "", fmt.Errorf("Error creating replicacontroler named %q with error %q", name, stderr.String())
	}

	var runReturn kubectlItem
	err = parseJSON(&runReturn, stdout.String())
	if err != nil {
		return "", err
	}
	uid := runReturn.Metadata.Uid
	return uid, nil
}

//  calling function needs to handle if the read is successful but the rc is dead or has no replicas
func ReadKubeRC(name string) (int, error) {
	get_replicacontrolers := exec.Command("kubectl", "get", "rc", name, "--output=json")
	var stdout, stderr bytes.Buffer
	get_replicacontrolers.Stdout = &stdout
	get_replicacontrolers.Stderr = &stderr
	err := get_replicacontrolers.Run()
	if err != nil {
		return -1, fmt.Errorf("Error listing replica controlers: %q", stderr.String())
	}

	var getReturn kubectlItem
	err = parseJSON(&getReturn, stdout.String())
	if err != nil {
		return -1, err
	}

	return getReturn.Status.Replicas, nil
}

func DeleteKubeRc(name string) (error) {
	delete_replicacontrolers := exec.Command("kubectl", "delete", "rc", name)
	var stdout, stderr bytes.Buffer
	delete_replicacontrolers.Stdout = &stdout
	delete_replicacontrolers.Stderr = &stderr
	err := delete_replicacontrolers.Run()
	if err != nil {
		return  fmt.Errorf("Error listing replica controlers: %q", stderr.String())
	}
	
	return nil
}
