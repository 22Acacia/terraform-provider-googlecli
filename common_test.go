package terraformGcloud

import (
	"testing"
)


type sj struct {
	Name	string	`json:"NaMe"`
	Height	int	`json:"TallAmount"`
}

func TestParseJSON(t *testing.T) {
	sample_json := "{\"NaMe\":\"Gaston\",\"TallAmount\":300}"
	var samjson sj
	err := parseJSON(&samjson, sample_json)
	if err != nil {
		t.Error("Error parsing sample json: " + err.Error())
	}

	if samjson.Height != 300 || samjson.Name != "Gaston" {
		t.Fatal("Failed to parse sample json correctly")
	}
}
