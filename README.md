# terraform-provider-googlecli
This provider exposes resources that are available through 
google provided CLIs.  These clis are wrapped for one of 
two reasons:
1. there is no go operational API to hit
2. expediency

The contents of this provider will change over time as 
google services stablize and become available in the upstream
google provider.  

To use:
- check out
- run tests 
  - add several variables to your environment:
    - TF_ACC (set to anything)
    - GOOGLE_CREDENTIALS, set to the contents of a secrets file downloaded from google
    - GOOGLE_PROJECT, set to the project for the above credentials
    - GOOGLE_REGION, set to us-central1
  - execute tests using makefile
    - make tests TESTARGS=<args to pass to 'go test'>
    - ex to only run dataflow tests "make test TESTARGS='--run=Dataflow'"
- install binary to $GOBIN to make it usable system wide (assumes GOBIN is in your PATH)
  - make install
- edit terraform.rc (see terraform docs here: https://terraform.io/docs/plugins/basics.html) to have the
  following block:
  providers {
    googlecli = "terraform-provider-googlecli"
  }
- build and copy file to terraform install
  - locate terraform install
  - make build
  - cp terraform-provider-googlecli TERRAFORM_INSTALL_LOCATION
