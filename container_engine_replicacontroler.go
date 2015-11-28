package main

import (
	"fmt"
	"bytes"
	"os/exec"
)

//  only modeling the parts of the responses as I need them
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

type kubectlService struct {
	Status		struct{
		Ingress	[]struct {
			IP	string 	`json:"ip"`
		} 			`json:"ingress"`
	} 				`json:"status"`
}

func addExtraArgs(run_args []string, optional_args, env_args map[string]string) ([]string) {
	for k, v := range optional_args {
		run_args = append(run_args, "--" + k + "=" +v)
	}
	for k, v := range env_args {
		run_args = append(run_args, "--env=\"" + k + "=" + v + "\"")
	}

	return run_args
}

func CreateKubeRC(name, dockerImage, external_port string, optional_args, env_args map[string]string) (string, error) {
	kubectl_cmd := "kubectl"
	kubectl_run_args :=[]string{"run", name, "--image=" + dockerImage, "--output=json"}
	kubectl_run_args = addExtraArgs(kubectl_run_args, optional_args, env_args)
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

	var ex_err error
	if external_port != "" {
		ex_err = expose_rc_externally(name, external_port)
	}

	return uid,ex_err
}

func expose_rc_externally(name, external_port string) (error) {
	expose_rc := exec.Command("kubectl", "expose", "rc", name, "--port=" + external_port, "--create-external-load-balancer=true", "--output=json")
	var stdout, stderr bytes.Buffer
	expose_rc.Stdout = &stdout
	expose_rc.Stderr = &stderr
	err := expose_rc.Run()
	if err != nil {
		return fmt.Errorf("Error creating external load balancer on a specific port: %q", stderr.String())
	}

	return nil
}

//  calling function needs to handle if the read is successful but the rc is dead or has no replicas
func ReadKubeRC(name, external_port string) (int, string, error) {
	get_replicacontrolers := exec.Command("kubectl", "get", "rc", name, "--output=json")
	var stdout, stderr bytes.Buffer
	get_replicacontrolers.Stdout = &stdout
	get_replicacontrolers.Stderr = &stderr
	err := get_replicacontrolers.Run()
	if err != nil {
		return -1, "", fmt.Errorf("Error listing replica controlers: %q", stderr.String())
	}

	var getReturn kubectlItem
	err = parseJSON(&getReturn, stdout.String())
	if err != nil {
		return -1, "", err
	}

	var ex_err error
	var external_ip string
	if external_port != "" {
		external_ip, ex_err = fetchExternalIp(name)
	}
	return getReturn.Status.Replicas, external_ip, ex_err
}

func fetchExternalIp(name string) (string, error) {
	get_services := exec.Command("kubectl", "get", "services", name, "--output=json")
	var stdout, stderr bytes.Buffer
	get_services.Stdout = &stdout
	get_services.Stderr = &stderr
	err := get_services.Run()
	if err != nil {
		return "", fmt.Errorf("Error fetching services information for this cluster: %q", stderr.String())
	}

	var getReturn kubectlService
	err = parseJSON(&getReturn, stdout.String())
	if err != nil {
		return "", err
	}

	var ex_ip string
	if len(getReturn.Status.Ingress) != 0 {
		ex_ip = getReturn.Status.Ingress[0].IP
	}

	return ex_ip, nil
}

func DeleteKubeRC(name string) (error) {
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
