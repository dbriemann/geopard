# geopard
A fast and slim geocoding library written in Go.

## Supported APIs
Currently only supports the Google geocoding api. See its [docs](https://developers.google.com/maps/documentation/geocoding/intro) for more information.

## How to install
	$ go get github.com/dbriemann/geopard

## Sample application
```Go
package main

import (
	"fmt"

	"github.com/dbriemann/geopard"
)

func main() {
	if loc, err := geopard.GoogleGeocode("New York"); err != nil {
		fmt.Println(err.Error())
	} else {
		//prints the formatted address for the queried location
		fmt.Println(loc.Results[0].FormattedAddr)
	}
}
```
