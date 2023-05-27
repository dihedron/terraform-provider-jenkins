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

## Usage

In order to create Jobs inside of your Jenkins installation, you need to describe them in your Terraform file (`.tf`).

Unfortunately Jenkins APIs lack abstraction and in order to create a Job you need to post a whole XML describing the Job itself and the configuration of the plugins you have in your installation. Thus, in order to create resources you need to either embed the XML inside the `.tf` file or provide the path to a file that contains the template.

You can check how you can do it in the `tests/` directory, where the `main.tf` provides both examples; the second one refers to a separate `job_template.xml` which is in the same directory.

### Where do I get the Job XML template?

Create a Project manually, then look for the XML file that was created on the Jenkins server; that file is your starting point.
You can replace some of the hardcoded values you provided during the manual configuration process with template variables (google "Golang templates" for an overview of how templating work in Go).
Again the `job_template.xml` file was produced like this, and then edited to "templatise" some of the values.

Thus, if you need to create three types of Jobs:

1. create them manually
2. retrieve the three XMLs from your Jenkins server
3. templatise and save them as `job1.xml`, `job2.xml` and `job3.xml` 
4. proceed with your Terraform files, referring back to the templates as neecessary (see the second job in `tests/main.tf` to see how.

