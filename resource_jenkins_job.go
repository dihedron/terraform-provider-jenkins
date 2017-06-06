package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	jenkins "github.com/bndr/gojenkins"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceJenkinsJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceJenkinsJobCreate,
		Read:   resourceJenkinsJobRead,
		Update: resourceJenkinsJobUpdate,
		Delete: resourceJenkinsJobDelete,
		Exists: resourceJenkinsJobExists,
		Schema: map[string]*schema.Schema{
			// this is the job's ID (primary key)
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The unique name of the JenkinsCI job.",
				Required:    true,
			},
			"display_name": &schema.Schema{
				Type: schema.TypeString,
				Description: "If set, the optional display name is shown for the job throughout the Jenkins web GUI; " +
					"it needs not be unique among all jobs, and defaults to the job name.",
				Optional: true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The (optional) description of the JenkinsCI job.",
				Optional:    true,
			},
			"disabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "When this option is checked, no new builds of this project will be executed.",
				Optional:    true,
			},
			"template": &schema.Schema{
				Type: schema.TypeString,
				Description: "The configuration file template; it can be provided inline (e.g. as an HEREDOC), as a web " +
					"URL pointing to a text file (http://... or https://...), or as filesystem URL (file://...).",
				Required: true,
			},
			"parameters": {
				Type:        schema.TypeMap,
				Description: "The set of parameters to be set in the template to generate a valid config.xml file.",
				Optional:    true,
				Elem:        schema.TypeString,
			},
			"hash": &schema.Schema{
				Type: schema.TypeString,
				Description: "This internal parameter keeps track of modifications to the template when it is not " +
					"embedded into the job configuration; the has is computed each time the status is refreshed and " +
					"compared with the value stored here, so that any change to the template can be detected.",
				Computed: true,
			},
		},
	}
}

func resourceJenkinsJobExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*jenkins.Jenkins)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] jenkins::exists - checking if job %q exists", name)

	_, err := client.GetJob(name)
	if err != nil {
		log.Printf("[DEBUG] jenkins::exists - job %q does not exist: %v", name, err)
		return false, nil
	}

	log.Printf("[DEBUG] jenkins::exists - job %q exists", name)
	return true, nil
}

func resourceJenkinsJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)

	name := d.Get("name").(string)
	xml, err := createConfigXML(d)
	if err != nil {
		return err
	}
	job, err := client.CreateJob(xml, name)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] jenkins::create - job %q created", name)

	d.SetId(job.GetName())
	return err
}

func resourceJenkinsJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] jenkins::read - looking for job %q", name)

	job, err := client.GetJob(name)
	if err != nil {
		log.Printf("[DEBUG] jenkins::read - job %q does not exist: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] jenkins::read - job %q exists", job.GetName())

	d.SetId(job.GetName())

	return nil
}

func resourceJenkinsJobUpdate(d *schema.ResourceData, meta interface{}) error {
	var name string
	client := meta.(*jenkins.Jenkins)

	d.Partial(true)

	var job *jenkins.Job
	var err error

	if d.HasChange("name") {
		old, new := d.GetChange("name")

		job, err = client.GetJob(old.(string))
		if err != nil {
			log.Printf("[ERROR] jenkins::update - error retrieving job %q", old.(string))
			return err
		}

		ok, err := job.Rename(new.(string))
		if !ok || err != nil {
			log.Printf("[ERROR] jenkins::update - error renaming job %q to %q: %v", old.(string), new.(string), err)
			return err
		}

		log.Printf("[DEBUG] jenkins::update - job %q renamed to %q", old.(string), new.(string))
		d.SetPartial("name")
		name = new.(string)
	} else {
		name = d.Get("name").(string)
	}

	// grab job by current name
	job, err = client.GetJob(name)

	// if the template is not embedded, retrieve it and compute its hash, then
	// compare it against the value that was computed the last time: it may have
	// changed on disk or on the remote web server
	value := d.Get("template").(string)
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "file://") {
		log.Printf("[DEBUG] jenkins::update - need to recompute hash for template")
	}

	// at this point job has been initialised
	if d.HasChange("display_name") || d.HasChange("description") || d.HasChange("disabled") || d.HasChange("template") || d.HasChange("parameters") {
		name := d.Get("name").(string)

		xml, err := createConfigXML(d)
		if err != nil {
			return err
		}

		err = job.UpdateConfig(xml)
		if err != nil {
			log.Printf("[ERROR] jenkins::update - error updating job %q configuration: %v", name, err)
			return err
		}
		d.SetPartial("display_name")
		d.SetPartial("description")
		d.SetPartial("disabled")
		d.SetPartial("template")
		d.SetPartial("parameters")
	}
	d.Partial(false)
	return nil
}

func resourceJenkinsJobDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)
	name := d.Id()

	log.Printf("[DEBUG] jenkins_pipeline::delete - removing %q", name)

	ok, err := client.DeleteJob(name)

	log.Printf("[DEBUG] jenkins_pipeline::delete - %q removed: %t", name, ok)
	return err
}

// Job contains all the data pertaining to a Jenkins job, in a format that is
// easy to use with Golang text/templates
type Job struct {
	Name        string
	Description string
	DisplayName string
	Disabled    bool
	Parameters  map[string]string
}

func createConfigXML(d *schema.ResourceData) (string, error) {
	var configuration string
	value, ok := d.GetOk("template")
	if !ok {
		log.Printf("[ERROR] jenkins::xml - invalid config.xml template")
		return "", fmt.Errorf("Invalid config.xml template")
	}

	// if necessary, download the config.xml template from the server
	configuration = value.(string)
	if strings.HasPrefix(configuration, "http://") || strings.HasPrefix(configuration, "https://") {
		log.Printf("[DEBUG] jenkins::xml - retrieving template from URL %q", configuration)
		response, err := http.Get(configuration)
		if err != nil {
			log.Printf("[ERROR] jenkins::xml - error connecting to HTTP server: %v", err)
			return "", err
		}
		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("[ERROR] jenkins::xml - error reading HTTP server response: %v", err)
			return "", err
		}
		configuration = string(data)
	} else if strings.HasPrefix(configuration, "file://") {
		log.Printf("[DEBUG] jenkins::xml - retrieving template from filesystem: %q", configuration)
		configuration = strings.Replace(configuration, "file://", "", 1)
		data, err := ioutil.ReadFile(configuration)
		if err != nil {
			log.Printf("[ERROR] jenkins::xml - error reading from filesystem: %v", err)
			return "", err
		}
		configuration = string(data)
	}
	log.Printf("[DEBUG] jenkins::xml - template:\n%s", configuration)

	// create and parse the config.xml template
	tpl, err := template.New("template").Parse(configuration)
	if err != nil {
		log.Printf("[ERROR] jenkins::xml - error parsing template: %v", err)
		return "", err
	}

	// now copy the input parameters into a data structure that is compatible
	// with the config.xml template
	job := &Job{
		Name:       d.Get("name").(string),
		Parameters: map[string]string{},
	}
	if value, ok := d.GetOk("display_name"); ok {
		job.DisplayName = value.(string)
	}
	if value, ok := d.GetOk("description"); ok {
		job.Description = value.(string)
	}
	if value, ok := d.GetOk("disabled"); ok {
		job.Disabled = value.(bool)
	}
	if value, ok := d.GetOk("parameters"); ok {
		value := value.(map[string]interface{})
		for k, v := range value {
			job.Parameters[k] = v.(string)
		}
	}

	// apply the job object to the template
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, job)
	if err != nil {
		log.Printf("[ERROR] jenkis::xml - error executing template: %v", err)
	}
	return buffer.String(), nil
}
