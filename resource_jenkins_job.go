package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

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
			// this is the job's ID (primary key)
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The unique name of the JenkinsCI job.",
				Required:    true,
				ForceNew:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The (optional) description of the JenkinsCI job.",
				Optional:    true,
				ForceNew:    true, // TODO:remove once update is available
			},
			"display_name": &schema.Schema{
				Type: schema.TypeString,
				Description: "If set, the optional display name is shown for the job throughout the Jenkins web GUI; " +
					"it needs not be unique among all jobs, and defaults to the job name.",
				Optional: true,
				ForceNew: true, // TODO:remove once update is available
			},
			"disabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "When this option is checked, no new builds of this project will be executed.",
				Optional:    true,
				ForceNew:    true, // TODO:remove
			},
			"template": &schema.Schema{
				Type: schema.TypeString,
				Description: "The configuration file template (e.g. as an HEREDOC) or a URL pointing to the same " +
					"file; if the string starts with an HTTP or HTTPS schema, the plugin will automatically try to download " +
					"it before applying the given parameters.",
				Required: true,
				ForceNew: true, // TODO:remove once update is available
			},
			"parameters": {
				Type:        schema.TypeMap,
				Description: "The set of parameters to be set in the template to generate a valid config.xml file.",
				Optional:    true,
				ForceNew:    true, // TODO: remove
				Elem:        schema.TypeString,
			},
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
		log.Printf("[DEBUG] jenkins::xml - retrieving template from %q", configuration)
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
	if value, ok := d.GetOk("description"); ok {
		job.Description = value.(string)
	}
	if value, ok := d.GetOk("display_name"); ok {
		job.DisplayName = value.(string)
	}
	if value, ok := d.GetOk("disabled"); ok {
		switch value := value.(type) {
		case bool:
			job.Disabled = value
		case string:
			disabled, err := strconv.ParseBool(value)
			if err == nil {
				job.Disabled = disabled
			}
		}
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
