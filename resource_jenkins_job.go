package main

import (
	"log"
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
					"embedded into the job configuration; the hash is computed each time the status is refreshed and " +
					"compared with the value stored here, so that any change to the template can be detected.",
				Optional: true,
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
		// TODO: check error when resource does not exist
		// remove from state
		d.SetId("")
		return false, nil
	}

	log.Printf("[DEBUG] jenkins::exists - job %q exists", name)
	return true, nil
}

func resourceJenkinsJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*jenkins.Jenkins)

	name := d.Get("name").(string)

	cxt, err := NewConfigXMLTemplate(d.Get("template").(string))
	if err != nil {
		log.Printf("[ERROR] jenkins::create - error creating config.xml template object for %q: %v", name, err)
		return err
	}

	sha, err := cxt.Hash()
	if err != nil {
		log.Printf("[ERROR] jenkins::create - error computing config.xml template hash for %q: %v", name, err)
		return err
	}

	xml, err := cxt.BindTo(d)
	if err != nil {
		log.Printf("[ERROR] jenkins::create - error binding config.xml template to %q: %v", name, err)
		return err
	}

	job, err := client.CreateJob(xml, name)
	if err != nil {
		log.Printf("[ERROR] jenkins::create - error creating job for %q: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] jenkins::create - job %q created", name)

	d.SetId(job.GetName())
	d.Set("hash", sha)
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

	cxt, err := NewConfigXMLTemplate(d.Get("template").(string))
	if err != nil {
		log.Printf("[ERROR] jenkins::read - error creating config.xml template object for %q: %v", name, err)
		return err
	}

	sha, err := cxt.Hash()
	if err != nil {
		log.Printf("[ERROR] jenkins::read - error computing config.xml template hash for %q: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] jenkins::read - job %q exists (hash was %q, now %q)", job.GetName(), d.Get("hash").(string), sha)

	d.SetId(job.GetName())
	//d.Set("hash", sha)

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

	// retrieve the template and compute its hash, then compare the new value
	// against the value that was computed the last time and saved into the
	// state: the template may have changed on disk or on the remote web server
	// even if its URL hasn't
	cxt, err := NewConfigXMLTemplate(d.Get("template").(string))
	if err != nil {
		log.Printf("[ERROR] jenkins::update - error creating config.xml template object for %q: %v", name, err)
		return err
	}

	sha, err := cxt.Hash()
	if err != nil {
		log.Printf("[ERROR] jenkins::update - error computing config.xml template hash for %q: %v", name, err)
		return err
	}

	templateChanged := !strings.EqualFold(d.Get("hash").(string), sha)
	log.Printf("[DEBUG] jenkins::update - the template hash has changed? %t", templateChanged)

	// at this point job has been initialised
	if d.HasChange("display_name") || d.HasChange("description") || d.HasChange("disabled") || d.HasChange("template") || d.HasChange("parameters") || d.HasChange("hash") || templateChanged {
		name := d.Get("name").(string)

		xml, err := cxt.BindTo(d)
		if err != nil {
			log.Printf("[ERROR] jenkins::update - error binding config.xml template to %q: %v", name, err)
			return err
		}

		err = job.UpdateConfig(xml)
		if err != nil {
			log.Printf("[ERROR] jenkins::update - error updating job %q configuration: %v", name, err)
			return err
		}
		// update the template hash in the state
		d.Set("hash", sha)
		d.SetPartial("hash")
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
