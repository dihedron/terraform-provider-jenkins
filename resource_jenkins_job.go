package main

import (
	"log"

	jenkins "github.com/bndr/gojenkins"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceJenkinsJob() *schema.Resource {
	return &schema.Resource{
		//Create: resourceJenkinsJobCreate,
		Read: resourceJenkinsJobRead,
		//Update: resourceJenkinsJobUpdate,
		//Delete: resourceJenkinsJobDelete,
		Exists: resourceJenkinsJobExists,

		/*
			Importer: &schema.ResourceImporter{
				State: resourceJenkinsJobImport,
			},
		*/
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the JenkinsCi job.",
				Required:    true,
				ForceNew:    true,
			},
			/*
				"object_classes": &schema.Schema{
					Type:        schema.TypeSet,
					Description: "The set of classes this object conforms to (e.g. organizationalUnit, inetOrgPerson).",
					Elem:        &schema.Schema{Type: schema.TypeString},
					Set:         schema.HashString,
					Required:    true,
				},
				"attributes": &schema.Schema{
					Type:        schema.TypeSet,
					Description: "The map of attributes of this object; each attribute can be multi-valued.",
					Set:         attributeHash,
					MinItems:    0,

					Elem: &schema.Schema{
						Type:        schema.TypeMap,
						Description: "The list of values for a given attribute.",
						MinItems:    1,
						MaxItems:    1,
						Elem: &schema.Schema{
							Type:        schema.TypeString,
							Description: "The individual value for the given attribute.",
						},
					},
					Optional: true,
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

/*
func resourceJenkinsJobCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)
	dn := d.Get("dn").(string)

	log.Printf("[DEBUG] ldap_object::create - creating a new object under %q", dn)

	request := ldap.NewAddRequest(dn)

	// retrieve classe from HCL
	objectClasses := []string{}
	for _, oc := range (d.Get("object_classes").(*schema.Set)).List() {
		log.Printf("[DEBUG] ldap_object::create - object %q has class: %q", dn, oc.(string))
		objectClasses = append(objectClasses, oc.(string))
	}
	request.Attribute("objectClass", objectClasses)

	// if there is a non empty list of attributes, loop though it and
	// create a new map collecting attribute names and its value(s); we need to
	// do this because we could not model the attributes as a map[string][]string
	// due to an appareent limitation in HCL; we have a []map[string]string, so
	// we loop through the list and accumulate values when they share the same
	// key, then we use these as attributes in the LDAP client.
	if v, ok := d.GetOk("attributes"); ok {
		attributes := v.(*schema.Set).List()
		if len(attributes) > 0 {
			log.Printf("[DEBUG] ldap_object::create - object %q has %d attributes", dn, len(attributes))
			m := make(map[string][]string)
			for _, attribute := range attributes {
				log.Printf("[DEBUG] ldap_object::create - %q has attribute of type %T", dn, attribute)
				// each map should only have one entry (see resource declaration)
				for name, value := range attribute.(map[string]interface{}) {
					log.Printf("[DEBUG] ldap_object::create - %q has attribute[%v] => %v (%T)", dn, name, value, value)
					m[name] = append(m[name], value.(string))
				}
			}
			// now loop through the map and add attributes with theys value(s)
			for name, values := range m {
				request.Attribute(name, values)
			}
		}
	}

	err := client.Add(request)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] ldap_object::create - object %q added to LDAP server", dn)

	d.SetId(dn)
	return resourceLDAPObjectRead(d, meta)
}
*/

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

func resourceLDAPObjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)
	dn := d.Get("dn").(string)

	log.Printf("[DEBUG] ldap_object::delete - removing %q", dn)

	request := ldap.NewDelRequest(dn, nil)

	err := client.Del(request)
	if err != nil {
		log.Printf("[ERROR] ldap_object::delete - error removing %q: %v", dn, err)
		return err
	}
	log.Printf("[DEBUG] ldap_object::delete - %q removed", dn)
	return nil
}

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
