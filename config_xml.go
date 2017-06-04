package main

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func createConfigXML(d *schema.ResourceData) string {
	var buffer bytes.Buffer
	buffer.WriteString("<?xml version='1.0' encoding='UTF-8'?>\n")
	buffer.WriteString("<flow-definition plugin=\"workflow-job@2.11\">\n")
	buffer.WriteString(" <actions/>\n")
	if value, ok := d.GetOk("description"); ok {
		buffer.WriteString(fmt.Sprintf(" <description>%s</description>\n", value.(string)))
	}
	if value, ok := d.GetOk("display_name"); ok {
		buffer.WriteString(fmt.Sprintf(" <displayName>%s</displayName>\n", value.(string)))
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
		buffer.WriteString("  <com.coravy.hudson.plugins.github.GithubProjectProperty plugin=\"github@1.27.0\">\n")
		buffer.WriteString(fmt.Sprintf("   <projectUrl>%s</projectUrl>\n", value["project_url"].(string)))
		buffer.WriteString(fmt.Sprintf("   <displayName>%s</displayName>\n", value["display_name"].(string)))
		buffer.WriteString("  </com.coravy.hudson.plugins.github.GithubProjectProperty>\n")
	}

	// TODO: parameters

	if value, ok := d.GetOk("throttle_builds"); ok {
		value := value.([]interface{})[0].(map[string]interface{})
		buffer.WriteString("  <jenkins.branch.RateLimitBranchProperty_-JobPropertyImpl plugin=\"branch-api@2.0.9\">\n")
		buffer.WriteString(fmt.Sprintf("   <durationName>%s</durationName>\n", value["period"].(string)))
		buffer.WriteString(fmt.Sprintf("   <count>%d</count>\n", value["rate"].(int)))
		buffer.WriteString("  </jenkins.branch.RateLimitBranchProperty_-JobPropertyImpl>\n")
	}

	// there are multiple parameters that can count as triggers
	_, ok1 := d.GetOk("build_after")
	_, ok2 := d.GetOk("periodic_build_schedule")
	_, ok3 := d.GetOk("github_hook_trigger")
	_, ok4 := d.GetOk("scm_poll_trigger")
	if ok1 || ok2 || ok3 || ok4 {
		buffer.WriteString("  <org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>\n")
		buffer.WriteString("   <triggers>\n")

		// 1. an upstream project can trigger a build
		if value, ok := d.GetOk("build_after"); ok {
			value := value.([]interface{})[0].(map[string]interface{})
			buffer.WriteString("    <jenkins.triggers.ReverseBuildTrigger>\n")
			buffer.WriteString("     <spec></spec>\n")
			buffer.WriteString(fmt.Sprintf("     <upstreamProjects>%s</upstreamProjects>\n", value["projects"].(string)))
			buffer.WriteString("     <threshold>\n")
			switch value["threshold"] {
			case "success":
				buffer.WriteString("      <name>SUCCESS</name>\n")
				buffer.WriteString("      <ordinal>0</ordinal>\n")
				buffer.WriteString("      <color>BLUE</color>\n")
				buffer.WriteString("      <completeBuild>true</completeBuild>\n")
			case "unstable":
				buffer.WriteString("      <name>UNSTABLE</name>\n")
				buffer.WriteString("      <ordinal>1</ordinal>\n")
				buffer.WriteString("      <color>YELLOW</color>\n")
				buffer.WriteString("      <completeBuild>true</completeBuild>\n")
			case "failure":
				buffer.WriteString("      <name>FAILURE</name>\n")
				buffer.WriteString("      <ordinal>2</ordinal>\n")
				buffer.WriteString("      <color>RED</color>\n")
				buffer.WriteString("      <completeBuild>true</completeBuild>\n")
			}
			buffer.WriteString("     </threshold>\n")
			buffer.WriteString("    </jenkins.triggers.ReverseBuildTrigger>\n")
			// TODO: whatever's left to add here...
		}

		// 2. a timer can trigger a build
		if value, ok := d.GetOk("periodic_build_schedule"); ok {
			value := value.(string)
			buffer.WriteString("    <hudson.triggers.TimerTrigger>\n")
			buffer.WriteString(fmt.Sprintf("     <spec>%s</spec>\n", value))
			buffer.WriteString("    </hudson.triggers.TimerTrigger>\n")
		}

		// 3. a webhook from GitHub can trigger the builds
		if value, ok := d.GetOk("github_hook_trigger"); ok {
			value := value.(bool)
			if value {
				buffer.WriteString("    <com.cloudbees.jenkins.GitHubPushTrigger plugin=\"github@1.27.0\">\n")
				buffer.WriteString("     <spec></spec>\n")
				buffer.WriteString("    </com.cloudbees.jenkins.GitHubPushTrigger>\n")
			}
		}

		// 4. an SCM poll can trigger a build
		if value, ok := d.GetOk("scm_poll_trigger"); ok {
			value := value.([]interface{})[0].(map[string]interface{})
			buffer.WriteString("    <hudson.triggers.SCMTrigger>\n")
			buffer.WriteString(fmt.Sprintf("     <spec>%s</spec>\n", value["schedule"].(string)))
			buffer.WriteString(fmt.Sprintf("     <ignorePostCommitHooks>%t</ignorePostCommitHooks>\n", value["ignore_postcommit_hooks"].(bool)))
			buffer.WriteString("    </hudson.triggers.SCMTrigger>\n")
		}

		buffer.WriteString("   </triggers>\n")
		buffer.WriteString("  </org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>\n")
	}

	buffer.WriteString(" </properties>\n")
	buffer.WriteString(" <definition class=\"org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition\" plugin=\"workflow-cps@2.32\">\n")
	buffer.WriteString("  <script></script>\n")
	buffer.WriteString("  <sandbox>true</sandbox>\n")
	buffer.WriteString(" </definition>\n")
	buffer.WriteString(" <triggers/>\n")

	if value, ok := d.GetOk("quiet_period"); ok {
		buffer.WriteString(fmt.Sprintf(" <quietPeriod>%d</quietPeriod>\n", value.(int)))
	}

	if value, ok := d.GetOk("remote_trigger_token"); ok {
		buffer.WriteString(fmt.Sprintf(" <authToken>%s</authToken>\n", value.(string)))
	}

	if value, ok := d.GetOk("disabled"); ok {
		buffer.WriteString(fmt.Sprintf(" <disabled>%t</disabled>\n", value.(bool)))
	} else {
		buffer.WriteString(" <disabled>false</disabled>\n")
	}

	buffer.WriteString("</flow-definition>\n")
	return buffer.String()
}
