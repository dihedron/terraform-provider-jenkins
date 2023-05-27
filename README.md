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

Unfortunately Jenkins APIs lack abstraction and in order to create a Job you need to post the whole XML file (config.xml) describing the Job. This can be done either by embedding the XML inside the `.tf` file or providing the path to a file that contains the Job template.

This provider allows to create a template Job descriptor in XML format with variables, and then specifying the values in the resource descriptor in the Terraform file: 

```hcl
resource "jenkins_job" "second" {
	name 				  			   = "Second"
	display_name					   = "The Second Project Display Name"
	description			  			   = "The second job is created from a file on disk"
	disabled						   = true
	parameters  					   = {
		KeepDependencies 			   = true,
		GitLabConnection			   = "http://gitlab.example.com/my-second-project/project.git",
		TriggerOnPush				   = true,
		TriggerOnMergeRequest		   = true,
		TriggerOpenMergeRequestOnPush  = "never",
		TriggerOnNoteRequest           = true,
		NoteRegex                      = "Jenkins please retry a build",
		CISkip                         = true,
		SkipWorkInProgressMergeRequest = true,
		SetBuildDescription            = true,
		BranchFilterType               = "All",
		SecretToken                    = "{AQAAABAAAAAQwt1GRY9q3ZVQO3gt3epgTsk5dMX+jSacfO7NOzm5Eyk=}",
		UserRemoteConfig			   = "https://gitlab.example.com/confmgmt/user-web.git",
		BranchSpec                     = "*/master",
		GenerateSubmoduleConfiguration = false,
	}
	template						   = "file://./job_template.xml"
}
```
The `parameters` section is a map whose keys can be arbitrarily defined to match those in the `temaplate`.

You can check the `tests/` directory, where the `main.tf` provides both an embedded template and an external one (a separate `job_template.xml` in the same directory).

### Where do I get the Job's XML template?

Create a Project manually, then look for the XML file that was created on the Jenkins server; that file is your starting point.

You can replace some of the hardcoded values you provided during the manual configuration process with template variables you can place in the `parameters` section. In order to learn how to work with "GOlang templates", just google it, there is plenty of documentation available. In a few words, if you place `MYVar = MyValue` in the `parameters` section, you can use `{{ .MyVar }}` and the provider will automatically replace it with `MyValue`.

The `job_template.xml` file was produced like this, and then edited to "templatise" some of the values.

