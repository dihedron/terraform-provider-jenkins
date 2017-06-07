package main

import (
	"encoding/xml"
	"os"

	"github.com/hashicorp/terraform/plugin"
)

func main() {
	if len(os.Args) > 1 {
		// editor mode
		file, _ := os.Open(os.Args[1])
		d := xml.NewDecoder(file)
		for {
			t, tokenErr := d.Token()
			if tokenErr != nil {
				// TODO: log error
			}
			switch t := t.(type) {
			case xml.StartElement:
				if t.Name.Space == "foo" && t.Name.Local == "bar" {
					/*
						var b bar
						if err := d.DecodeElement(&b, &t); err != nil {
							// handle error
						}
					*/
					// do something with b
				}
				// TODO: add other cases
			}
		}

	} else {
		// terraform plugin mode
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: Provider,
		})
	}
}
