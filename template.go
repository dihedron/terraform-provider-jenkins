package main

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"encoding/hex"

	"github.com/hashicorp/terraform/helper/schema"
)

// ConfigXMLTemplate represents a config.xml template as an object.
type ConfigXMLTemplate struct {
	template string
}

func (c *ConfigXMLTemplate) String() string {
	if c == nil {
		return ""
	}
	return c.template
}

// NewConfigXMLTemplate creates a new ConfigXMLTemplate using the provided
// address or inline/embedded data.
func NewConfigXMLTemplate(from string) (*ConfigXMLTemplate, error) {

	config := &ConfigXMLTemplate{}

	if strings.HasPrefix(from, "http://") || strings.HasPrefix(from, "https://") {
		log.Printf("[DEBUG] jenkins::xml - retrieving template from URL %q", from)
		response, err := http.Get(from)
		if err != nil {
			log.Printf("[ERROR] jenkins::xml - error connecting to HTTP server: %v", err)
			return nil, err
		}
		defer response.Body.Close()
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("[ERROR] jenkins::xml - error reading HTTP server response: %v", err)
			return nil, err
		}
		config.template = string(data)
	} else if strings.HasPrefix(from, "file://") {
		log.Printf("[DEBUG] jenkins::xml - retrieving template from filesystem: %q", from)
		from = strings.Replace(from, "file://", "", 1)
		data, err := ioutil.ReadFile(from)
		if err != nil {
			log.Printf("[ERROR] jenkins::xml - error reading from filesystem: %v", err)
			return nil, err
		}
		config.template = string(data)
	} else {
		log.Printf("[DEBUG] jenkins::xml - template is inline: %q", from)
		config.template = from
	}

	return config, nil
}

// Hash returns the SHA-256 hash of the current (unbound) template.
func (c *ConfigXMLTemplate) Hash() (string, error) {
	if c == nil {
		log.Printf("[ERROR] jenkins::xml - invalid config.xml template object")
		return "", fmt.Errorf("Invalid config.xml template object")
	}
	hash := sha512.Sum512([]byte(c.template))
	return strings.ToLower(hex.EncodeToString(hash[:])), nil
}

// BindTo binds the current config.xml template to the given resource data.
func (c *ConfigXMLTemplate) BindTo(d *schema.ResourceData) (string, error) {

	if c == nil {
		log.Printf("[ERROR] jenkins::xml - invalid config.xml template object")
		return "", fmt.Errorf("Invalid config.xml template object")
	}
	log.Printf("[DEBUG] jenkins::xml - binding template:\n%s", c.template)

	// create and parse the config.xml template
	tpl, err := template.New("template").Parse(c.template)
	if err != nil {
		log.Printf("[ERROR] jenkins::xml - error parsing template: %v", err)
		return "", err
	}

	// Job contains all the data pertaining to a Jenkins job, in a format that is
	// easy to use with Golang text/templates
	type job struct {
		Name        string
		Description string
		DisplayName string
		Disabled    bool
		Parameters  map[string]string
	}

	// now copy the input parameters into a data structure that is compatible
	// with the config.xml template
	j := &job{
		Name:       d.Get("name").(string),
		Parameters: map[string]string{},
	}
	if value, ok := d.GetOk("display_name"); ok {
		j.DisplayName = value.(string)
	}
	if value, ok := d.GetOk("description"); ok {
		j.Description = value.(string)
	}
	if value, ok := d.GetOk("disabled"); ok {
		j.Disabled = value.(bool)
	}
	if value, ok := d.GetOk("parameters"); ok {
		value := value.(map[string]interface{})
		for k, v := range value {
			j.Parameters[k] = v.(string)
		}
	}

	// apply the job object to the template
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, j)
	if err != nil {
		log.Printf("[ERROR] jenkis::xml - error executing template: %v", err)
		return "", err
	}
	return buffer.String(), nil
}
