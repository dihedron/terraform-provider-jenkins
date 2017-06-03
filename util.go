package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func validateAllowedStringsCaseInsensitive(ss []string) schema.SchemaValidateFunc {
	log.Printf("[DEBUG] jenkins_pipeline::validate - length of %v: %d", ss, len(ss))
	ll := make([]string, 0, len(ss))
	for _, s := range ss {
		ll = append(ll, strings.ToLower(s))
	}
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := strings.ToLower(v.(string))
		existed := false
		for _, s := range ll {
			if strings.ToLower(s) == value {
				existed = true
				break
			}
		}
		if !existed {
			errors = append(errors, fmt.Errorf(
				"%q must contain a valid string value should in array %#v, got %q",
				k, ll, value))
		}
		return

	}
}
