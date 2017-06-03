package main

import (
	"log"

	jenkins "github.com/bndr/gojenkins"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceJenkinsJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceJenkinsJobCreate,
		Read:   resourceJenkinsJobRead,
		//Update: resourceJenkinsJobUpdate,
		Delete: resourceJenkinsJobDelete,
		Exists: resourceJenkinsJobExists,

		/*
			Importer: &schema.ResourceImporter{
				State: resourceJenkinsJobImport,
			},
		*/
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				// this is the job's ID
				Type:        schema.TypeString,
				Description: "The name of the JenkinsCI job.",
				Required:    true,
				ForceNew:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The (optional) description of the JenkinsCI job.",
				Optional:    true,
				ForceNew:    true, // TODO:remove once update is available
				//
				// goes into:
				// <description>This is the pipeline-archetype description</description>
				//
			},
			"build_discard_policy": {
				Type: schema.TypeList,
				Description: "Determines when, if ever, build records for this project should be discarded. Build records " +
					"include the console output, archived artifacts, and any other metadata related to a particular build.",
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				ForceNew: true, // TODO: remove
				//PromoteSingle: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"days_to_keep_builds": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "If not empty, build records are only kept up to this number of days.",
							Default:     0,
							Optional:    true,
							ForceNew:    true, // TODO:remove
						},
						"max_n_of_builds_to_keep": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "If not empty, only up to this number of build records are kept.",
							Default:     0,
							Optional:    true,
							ForceNew:    true, // TODO:remove
						},
						"days_to_keep_artifacts": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "If not empty, artifacts from builds older than this number of days will be deleted, but the logs, history, reports, etc for the build will be kept.",
							Default:     0,
							Optional:    true,
							ForceNew:    true,
						},
						"max_n_of_artifacts_to_keep": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "If not empty, only up to this number of builds have their artifacts retained.",
							Default:     0,
							Optional:    true,
							ForceNew:    true, // TODO:remove
						},
					},
					//
					// goes into:
					//    <jenkins.model.BuildDiscarderProperty>
					//       <strategy class="hudson.tasks.LogRotator">
					//         <daysToKeep>1</daysToKeep>
					//         <numToKeep>2</numToKeep>
					//         <artifactDaysToKeep>3</artifactDaysToKeep>
					//         <artifactNumToKeep>4</artifactNumToKeep>
					//       </strategy>
					//     </jenkins.model.BuildDiscarderProperty>
					// along with the following four parameters.
					//
				},
			},

			"disallow_concurrent_builds": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Determines when if multiple builds of the project can be run in parallel.",
				Default:     false,
				Optional:    true,
				ForceNew:    true, // TODO:remove
				//
				// goes into:
				// <org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty/>
				//
			},
			"github_project": {
				Type:        schema.TypeList,
				Description: "If present, indicates that the project sources are hosted on GitHub at the given URL.",
				Optional:    true,
				ForceNew:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_url": &schema.Schema{
							Type: schema.TypeString,
							Description: "The URL for the GitHub hosted project (without the tree/master or tree/branch part); " +
								"or example: http://github.com/rails/rails for the Rails project.",
							Required: true,
							ForceNew: true, // TODO:remove
						},
						"display_name": &schema.Schema{
							Type: schema.TypeString,
							Description: "The context name for commit status if status builder or status publisher is defined " +
								"for this project; it should be small and clear; if empty, the job name will be used for builder " +
								"and publisher.",
							Optional: true,
							ForceNew: true, // TODO:remove
						},
					},
					//
					// goes into:
					//	  <com.coravy.hudson.plugins.github.GithubProjectProperty plugin="github@1.27.0">
					//	    <projectUrl>http://github.com/dihedron/libjpp/</projectUrl>
					//	    <displayName>Display name!!!</displayName>
					//	  </com.coravy.hudson.plugins.github.GithubProjectProperty>
					//
				},
			},
			/*
				"parameters": {
					Type: schema.TypeMap,
					Description: "Parameters are used to prompt users for one or more inputs that will be passed into the build; " +
						"each parameter has a Name and a Value, depending on the type, and is exported as environment variables " +
						"when the build starts, to be accessed as ${PARAMETER_NAME} (or %PARAMETER_NAME% under Windows).",
					Optional: true,
					ForceNew: true, // TODO: remove
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Type:        schema.TypeString,
								Description: "The default value of the field, which allows the user to save typing the actual value.",
								Required:    true,
								ForceNew:    true, // TODO: remove
								ValidateFunc: validateAllowedStringsCaseInsensitive([]string{
									"boolean", "choice", "credentials", "file", "list_subversion_tags",
									"multiline_strings", "password", "run", "string",
								}),
							},
							"default_value": {
								Type:        schema.TypeString,
								Description: "The default value of the field, which allows the user to save typing the actual value.",
								Optional:    true,
								ForceNew:    true, // TODO: remove
							},
							"description": {
								Type:        schema.TypeString,
								Description: "A description that will be shown to the user when asking for input.",
								Optional:    true,
								ForceNew:    true, // TODO: remove
							},
							// these go into:
							//	<hudson.model.ParametersDefinitionProperty>
							//      <parameterDefinitions>
							//        <hudson.model.PasswordParameterDefinition>
							//          <name>PASSWORD</name>
							//          <description>Password description.</description>
							//          <defaultValue>{AQAAABAAAAAQNyJ6xxd3E8vPilhXqThfsFU1glQFcw9g+jKFeGHCOXU=}</defaultValue>
							//        </hudson.model.PasswordParameterDefinition>
							//        <hudson.model.BooleanParameterDefinition>
							//          <name>BOOLEAN</name>
							//          <description>Boolean description</description>
							//          <defaultValue>true</defaultValue>
							//        </hudson.model.BooleanParameterDefinition>
							//        <hudson.model.ChoiceParameterDefinition>
							//          <name>CHOICE</name>
							//          <description>Choice description</description>
							//          <choices class="java.util.Arrays$ArrayList">
							//            <a class="string-array">
							//              <string>First</string>
							//              <string>Second</string>
							//              <string>Third</string>
							//            </a>
							//          </choices>
							//        </hudson.model.ChoiceParameterDefinition>
							//        <com.cloudbees.plugins.credentials.CredentialsParameterDefinition plugin="credentials@2.1.13">
							//          <name>CREDENTIALS</name>
							//          <description>Credentials description</description>
							//          <defaultValue>USERNAME_ID</defaultValue>
							//          <credentialType>com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey</credentialType>
							//          <required>true</required>
							//        </com.cloudbees.plugins.credentials.CredentialsParameterDefinition>
							//        <hudson.model.FileParameterDefinition>
							//          <name>File location</name>
							//          <description>File description</description>
							//        </hudson.model.FileParameterDefinition>
							//        <hudson.scm.listtagsparameter.ListSubversionTagsParameterDefinition plugin="subversion@2.7.2">
							//          <name>SUBVERSION</name>
							//          <description>Select a Subversion entry</description>
							//          <tagsDir>https://svn.alfresco.com</tagsDir>
							//          <credentialsId>USERNAME_ID</credentialsId>
							//          <tagsFilter>^master$</tagsFilter>
							//          <reverseByDate>true</reverseByDate>
							//          <reverseByName>true</reverseByName>
							//          <defaultValue>Default value</defaultValue>
							//          <maxTags>99</maxTags>
							//        </hudson.scm.listtagsparameter.ListSubversionTagsParameterDefinition>
							//        <hudson.model.TextParameterDefinition>
							//          <name>MULTILINE</name>
							//          <description>Multiline description</description>
							//          <defaultValue>Default multiline</defaultValue>
							//        </hudson.model.TextParameterDefinition>
							//        <hudson.model.PasswordParameterDefinition>
							//          <name>PASSWORD</name>
							//          <description>Password description</description>
							//          <defaultValue>{AQAAABAAAAAg/3sWZBb7pUTXQyO0jcPFc3R9MMJ/x7H0u38Ug6MKu0Kajfb2rJ7C6k/bJlKIbB/Z}</defaultValue>
							//        </hudson.model.PasswordParameterDefinition>
							//        <hudson.model.RunParameterDefinition>
							//          <name>RUN</name>
							//          <description>RUN Description</description>
							//          <projectName>RUN_PROJECT</projectName>
							//          <filter>COMPLETED</filter>
							//        </hudson.model.RunParameterDefinition>
							//        <hudson.model.StringParameterDefinition>
							//          <name>STRING</name>
							//          <description>String Description</description>
							//          <defaultValue>Default String</defaultValue>
							//        </hudson.model.StringParameterDefinition>
							//      </parameterDefinitions>
							//    </hudson.model.ParametersDefinitionProperty>					},
							//
						},
					},
				},
			*/
			"throttle_builds": {
				Type:        schema.TypeList,
				Description: "Enforces a minimum time between builds based on the desired maximum rate.",
				Optional:    true,
				ForceNew:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rate": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "The maximum number of builds allowed within the specified time period.",
							Optional:    true,
							ForceNew:    true, // TODO:remove
							Default:     1,
						},
						"period": &schema.Schema{
							Type:        schema.TypeString,
							Description: "The time period within which the rate will be enforced (e.g. 2 builds per hour).",
							Required:    true,
							ForceNew:    true, // TODO:remove
							ValidateFunc: validateAllowedStringsCaseInsensitive([]string{
								"hour", "day", "week", "month", "year",
							}),
						},
					},
					//
					// goes into:
					//    <jenkins.branch.RateLimitBranchProperty_-JobPropertyImpl plugin="branch-api@2.0.9">
					//      <durationName>week</durationName>
					//      <count>2</count>
					//    </jenkins.branch.RateLimitBranchProperty_-JobPropertyImpl>
					//
				},
			},

			"build_after": {
				Type:        schema.TypeList,
				Description: "The trigger so that when some other projects finish building, a new build is scheduled.",
				Optional:    true,
				ForceNew:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"projects": &schema.Schema{
							Type:        schema.TypeString,
							Description: "The names of the projects to watch.",
							Required:    true,
							ForceNew:    true, // TODO:remove
						},
						"threshold": &schema.Schema{
							Type:        schema.TypeString,
							Description: "Condition under which the build is triggered (stable only, even unstable, even failed).",
							Required:    true,
							ForceNew:    true, // TODO:remove
							ValidateFunc: validateAllowedStringsCaseInsensitive([]string{
								"success", "unstable", "failure",
							}),
						},
					},
					//
					// goes into:
					//    <org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
					//      <triggers>
					//        <jenkins.triggers.ReverseBuildTrigger>
					//          <spec></spec>
					//          <upstreamProjects>UPSTREAM_PROJECT</upstreamProjects>
					//          <threshold>
					//            <name>SUCCESS|UNSTABLE|FAILURE</name>
					//            <ordinal>0|1|2</ordinal>
					//            <color>BLUE|YELLOW|RED</color>
					//            <completeBuild>true</completeBuild>
					//          </threshold>
					//        </jenkins.triggers.ReverseBuildTrigger>
					//        [...]
					//      </triggers>
					//    </org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
					//
				},
			},
			/*
				"periodic_build_schedule": &schema.Schema{
					Type:        schema.TypeList,
					Description: "Determines the schedule of periodic builds in a cron-like format.",
					Optional:    true,
					ForceNew:    true, // TODO:remove
					//
					// goes into:
					//    <org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
					//      <triggers>
					//        [...]
					//        <jenkins.triggers.ReverseBuildTrigger/>
					//        <hudson.triggers.TimerTrigger>
					//          <spec># every fifteen minutes (perhaps at :07, :22, :37, :52)
					//H/15 * * * *
					//# every ten minutes in the first half of every hour (three times, perhaps at :04, :14, :24)
					//H(0-29)/10 * * * *</spec>
					//        </hudson.triggers.TimerTrigger>
					//      </triggers>
					//    </org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
					//
				},

					"github_hook_trigger": &schema.Schema{
						Type:        schema.TypeBool,
						Description: "Upon a PUSH request from the GitHub SCM hook, Jenkins will trigger Git polling.",
						Optional:    true,
						ForceNew:    true, // TODO:remove
						Default:     false,
						// goes into:
						//  <org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
						//      <triggers>
						//      [...]
						//        <com.cloudbees.jenkins.GitHubPushTrigger plugin="github@1.27.0">
						//          <spec></spec>
						//        </com.cloudbees.jenkins.GitHubPushTrigger>
						//      </triggers>
						//    </org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
						//
					},
					"scm_poll_schedule": &schema.Schema{
						Type:        schema.TypeList,
						Description: "Determines the schedule of SCM polling in a cron-like format.",
						Optional:    true,
						ForceNew:    true, // TODO:remove
						//
						// goes into:
						//  <org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
						//      <triggers>
						//      [...]
						//        <hudson.triggers.SCMTrigger>
						//          <spec># once every two hours at 45 minutes past the hour starting at 9:45 AM and finishing at 3:45 PM every weekday.
						//45 9-16/2 * * 1-5
						//# once in every two hours slot between 9 AM and 5 PM every weekday (perhaps at 10:38 AM, 12:38 PM, 2:38 PM, 4:38 PM)
						//H H(9-16)/2 * * 1-5</spec>
						//          <ignorePostCommitHooks>false|true</ignorePostCommitHooks>
						//        </hudson.triggers.SCMTrigger>
						//      </triggers>
						//    </org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty>
						//
					},
			*/
		},
	}
}

/*
func resourceLDAPObjectImport(d *schema.ResourceData, meta interface{}) (imported []*schema.ResourceData, err error) {
	err = resourceLDAPObjectRead(d, meta)
	imported = append(imported, d)
	return
}
*/

func resourceJenkinsJobExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*jenkins.Jenkins)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] jenkins_job::exists - checking if job %q exists", name)

	_, err := client.GetJob(name)
	if err != nil {
		log.Printf("[DEBUG] jenkins_job::exists - job %q does not exist: %v", name, err)
		return false, nil
	}

	log.Printf("[DEBUG] jenkins_job::exists - job %q exists", name)
	return true, nil
}

func resourceJenkinsJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)

	name := d.Get("name").(string)
	xml := createConfigXML(d)
	job, err := client.CreateJob(xml, name)

	log.Printf("[DEBUG] jenkins_pipeline::create - job %q created", name)

	d.SetId(job.GetName())
	return err //resourceLDAPObjectRead(d, meta)
}

func resourceJenkinsJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] jenkins_job::read - looking for job %q", name)

	job, err := client.GetJob(name)
	if err != nil {
		log.Printf("[DEBUG] jenkins_job::read - job %q does not exist: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] jenkins_job::read - job %q exists", job.GetName())

	d.SetId(job.GetName())
	//d.Set("name", job.GetName())

	return nil
}

/*
func resourceLDAPObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)

	log.Printf("[DEBUG] ldap_object::update - performing update on %q", d.Id())

	request := ldap.NewModifyRequest(d.Id())

	// handle objectClasses
	if d.HasChange("object_classes") {
		classes := []string{}
		for _, oc := range (d.Get("object_classes").(*schema.Set)).List() {
			classes = append(classes, oc.(string))
		}
		log.Printf("[DEBUG] ldap_object::update - updating classes of %q, new value: %v", d.Id(), classes)
		request.ReplaceAttributes = []ldap.PartialAttribute{
			ldap.PartialAttribute{
				Type: "objectClass",
				Vals: classes,
			},
		}
	}

	if d.HasChange("attributes") {

		o, n := d.GetChange("attributes")
		log.Printf("[DEBUG] ldap_object::update - \n%s", printAttributes("old attributes map", o))
		log.Printf("[DEBUG] ldap_object::update - \n%s", printAttributes("new attributes map", n))

		added, changed, removed := computeDeltas(o.(*schema.Set), n.(*schema.Set))
		if len(added) > 0 {
			log.Printf("[DEBUG] ldap_object::update - %d attributes added", len(added))
			request.AddAttributes = added
		}
		if len(changed) > 0 {
			log.Printf("[DEBUG] ldap_object::update - %d attributes changed", len(changed))
			if request.ReplaceAttributes == nil {
				request.ReplaceAttributes = changed
			} else {
				request.ReplaceAttributes = append(request.ReplaceAttributes, changed...)
			}
		}
		if len(removed) > 0 {
			log.Printf("[DEBUG] ldap_object::update - %d attributes removed", len(removed))
			request.DeleteAttributes = removed
		}
	}

	err := client.Modify(request)
	if err != nil {
		log.Printf("[ERROR] ldap_object::update - error modifying LDAP object %q with values %v", d.Id(), err)
		return err
	}
	return resourceLDAPObjectRead(d, meta)
}
*/

func resourceJenkinsJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)
	name := d.Id()

	log.Printf("[DEBUG] jenkins_pipeline::delete - removing %q", name)

	ok, err := client.DeleteJob(name)

	log.Printf("[DEBUG] jenkins_pipeline::delete - %q removed: %t", name, ok)
	return err
}

/*
// computes the hash of the map representing an attribute in the attributes set
func attributeHash(v interface{}) int {
	m := v.(map[string]interface{})
	var buffer bytes.Buffer
	buffer.WriteString("map {")
	for k, v := range m {
		buffer.WriteString(fmt.Sprintf("%q := %q;", k, v.(string)))
	}
	buffer.WriteRune('}')
	text := buffer.String()
	hash := hashcode.String(text)
	return hash
}

func printAttributes(prefix string, attributes interface{}) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s: {\n", prefix))
	if attributes, ok := attributes.(*schema.Set); ok {
		for _, attribute := range attributes.List() {
			for k, v := range attribute.(map[string]interface{}) {
				buffer.WriteString(fmt.Sprintf("    %q: %q\n", k, v.(string)))
			}
		}
		buffer.WriteRune('}')
	}
	return buffer.String()
}

func computeDeltas(os, ns *schema.Set) (added, changed, removed []ldap.PartialAttribute) {

	rk := NewSet() // names of removed attributes
	for _, v := range os.Difference(ns).List() {
		for k := range v.(map[string]interface{}) {
			rk.Add(k)
		}
	}

	ak := NewSet() // names of added attributes
	for _, v := range ns.Difference(os).List() {
		for k := range v.(map[string]interface{}) {
			ak.Add(k)
		}
	}

	kk := NewSet() // names of kept attributes
	for _, v := range ns.Intersection(os).List() {
		for k := range v.(map[string]interface{}) {
			kk.Add(k)
		}
	}

	ck := NewSet() // names of changed attributes

	// loop over remove attributes' names
	for _, k := range rk.List() {
		if !ak.Contains(k) && !kk.Contains(k) {
			// one value under this name has been removed, no other value has
			// been added back, and there is no further value under the same
			// name among those that were untouched; this means that it has
			// been dropped and must go among the RemovedAttributes
			log.Printf("[DEBUG} ldap_object::deltas - dropping attribute %q", k)
			removed = append(removed, ldap.PartialAttribute{
				Type: k,
				Vals: []string{},
			})
		} else {
			ck.Add(k)
		}
	}

	for _, k := range ak.List() {
		if !rk.Contains(k) && !kk.Contains(k) {
			// this is the first value under this name: no value is being
			// removed and no value is being kept; so we're adding this new
			// attribute to the LDAP object (AddedAttributes), getting all
			// the values under this name from the new set
			values := []string{}
			for _, m := range ns.List() {
				for mk, mv := range m.(map[string]interface{}) {
					if k == mk {
						values = append(values, mv.(string))
					}
				}
			}
			added = append(added, ldap.PartialAttribute{
				Type: k,
				Vals: values,
			})
			log.Printf("[DEBUG} ldap_object::deltas - adding new attribute %q with values %v", k, values)
		} else {
			ck.Add(k)
		}
	}

	// now loop over changed attributes and
	for _, k := range ck.List() {
		// the attributes in this set have been changed, in that a new value has
		// been added or removed and it was not the last/first one; so we're
		// adding this new attribute to the LDAP object (ModifiedAttributes),
		// getting all the values under this name from the new set
		values := []string{}
		for _, m := range ns.List() {
			for mk, mv := range m.(map[string]interface{}) {
				if k == mk {
					values = append(values, mv.(string))
				}
			}
		}
		changed = append(added, ldap.PartialAttribute{
			Type: k,
			Vals: values,
		})
		log.Printf("[DEBUG} ldap_object::deltas - changing attribute %q with values %v", k, values)
	}
	return
}
*/
