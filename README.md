# Terraform Jenkins 

[![CircleCI](https://circleci.com/gh/dihedron/terraform-provider-jenkins.svg?style=svg)](https://circleci.com/gh/dihedron/terraform-provider-jenkins)

## Installation

You can easily install the latest version with the following :

```
go get -u github.com/dihedron/terraform-provider-jenkins
```

Then add the plugin to your local `.terraformrc` :

```
cat >> ~/.terraformrc <<EOF
providers {
    jenkins = "${GOPATH}/bin/terraform-provider-jenkins"
}
EOF
```

