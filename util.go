package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func validateAllowedStringsCaseInsensitive(ss []string) schema.SchemaValidateFunc {
	ll := make([]string, 0, len(ss))
	for i, s := range ss {
		ll[i] = strings.ToLower(s)
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
