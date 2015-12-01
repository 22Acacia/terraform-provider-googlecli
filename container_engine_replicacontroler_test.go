package main

import (
	"testing"
)

func TestAddExtraArgs(t *testing.T) {
	run_args := []string{"Sound", "Of", "the" ,"Train"}
	optional_args := map[string]string{"Choo": "Choo"}
	env_args := map[string]string{"Green": "Leavesof"}

	run_args = addExtraArgs(run_args, optional_args, env_args)

	for idx, val := range []string{"Sound", "Of", "the", "Train", "--Choo=Choo", "--env=\"Green=Leavesof\""} {
		if val != run_args[idx] {
			t.Fatalf("Failed to create run_args list correctly.  Error occured at position %d where answer key is %s and generated value is %s. Generated %q instead of Sound Of the Train --Choo=Choo --env=\"Green=Leavesof\"", idx, val, run_args[idx], run_args)
		}
	}
}
