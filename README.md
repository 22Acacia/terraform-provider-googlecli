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
- build
- copy binary to somewhere in your path
- edit terraform.rc (see terraform docs for where) to have the
  following block:
  providers {
    googlecli = "terraform-provider-googlecli"
  }

