package main

import (
	"testing"
)

func TestExtractJobIds(t *testing.T) {
	sample_creation_return := "Dataflow SDK version: 1.1.1-SNAPSHOT\nSubmitted job: 2015-11-12_21_49_42-11275957002527739090"
	jobids := findJobIds(sample_creation_return)

	if len(jobids) != 1 && jobids[0] != "2015-11-12_21_49_42-11275957002527739090" {
		t.Fatal("Failed to extract jobids")
	}
}
