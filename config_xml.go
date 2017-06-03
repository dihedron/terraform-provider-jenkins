package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func createXML(d *schema.ResourceData) string {
	var buffer bytes.Buffer
	buffer.WriteString("<?xml version='1.0' encoding='UTF-8'?>\n")
	buffer.WriteString("<flow-definition plugin=\"workflow-job@2.11\">\n")
	buffer.WriteString(" <actions/>\n")
	if value, ok := d.GetOk("description"); ok {
		buffer.WriteString(fmt.Sprintf(" <description>%s</description>\n", value.(string)))
	}
	// this value is there but I don't know why...
	buffer.WriteString(" <keepDependencies>false</keepDependencies>\n")
	buffer.WriteString(" <properties>\n")
	if value, ok := d.GetOk("build_discard_policy"); ok {
		value := value.([]interface{})[0].(map[string]interface{})
		buffer.WriteString("  <jenkins.model.BuildDiscarderProperty>\n")
		buffer.WriteString("   <strategy class=\"hudson.tasks.LogRotator\">\n")
		buffer.WriteString(fmt.Sprintf("   <daysToKeep>%d</daysToKeep>\n", value["days_to_keep_builds"].(int)))
		buffer.WriteString(fmt.Sprintf("   <numToKeep>%d</numToKeep>\n", value["max_n_of_builds_to_keep"].(int)))
		buffer.WriteString(fmt.Sprintf("   <artifactDaysToKeep>%d</artifactDaysToKeep>\n", value["days_to_keep_artifacts"].(int)))
		buffer.WriteString(fmt.Sprintf("   <artifactNumToKeep>%d</artifactNumToKeep>\n", value["max_n_of_artifacts_to_keep"].(int)))
		buffer.WriteString("   </strategy>\n")
		buffer.WriteString("  </jenkins.model.BuildDiscarderProperty>\n")
	}

	if value, ok := d.GetOk("disallow_concurrent_builds"); ok {
		value := value.(bool)
		if value {
			buffer.WriteString("  <org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty/>\n")
		}
	}

	if value, ok := d.GetOk("github_project"); ok {
		value := value.([]interface{})[0].(map[string]interface{})
		log.Printf("[DEBUG] jenkins_pipeline::xml - type of value: %T", value)
		buffer.WriteString("  <com.coravy.hudson.plugins.github.GithubProjectProperty plugin=\"github@1.27.0\">\n")
		buffer.WriteString(fmt.Sprintf("   <projectUrl>%s</projectUrl>>\n", value["project_url"].(string)))
		buffer.WriteString(fmt.Sprintf("   <displayName>%s</displayName>\n", value["display_name"].(string)))
		buffer.WriteString("  </com.coravy.hudson.plugins.github.GithubProjectProperty>\n")
	}

	buffer.WriteString(" </properties>\n")
	buffer.WriteString(" <definition class=\"org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition\" plugin=\"workflow-cps@2.32\">\n")
	buffer.WriteString("  <script></script>\n")
	buffer.WriteString("  <sandbox>true</sandbox>\n")
	buffer.WriteString(" </definition>\n")
	buffer.WriteString(" <triggers/>\n")
	buffer.WriteString(" <disabled>false</disabled>\n")
	buffer.WriteString("</flow-definition>\n")
	return buffer.String()
}
