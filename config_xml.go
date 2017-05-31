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
		value := value.([]interface{})[0]
		log.Printf("[DEBUG] jenkins_pipeline::xml - type of value: %T", value)
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
