# geopard
## A fast and slim geocoding library written in Go.

### Supported APIs
Currently only supports the Google geocoding api. See its [docs](https://developers.google.com/maps/documentation/geocoding/intro) for more information.

### How to install
	$ go get github.com/dbriemann/geopard

### Examples
The 'hello world' of geopard would look like this
```Go
package main

import (
	"fmt"

	"github.com/dbriemann/geopard"
)

func main() {
	//create a singleton instance of geopard with default settings
	instance := geopard.GetInstance()

	if loc, err := instance.Geocode("New York"); err != nil {
		fmt.Println(err.Error())
	} else {
		addr := loc.Results[0]
		//prints the formatted address for the queried location
		fmt.Println(addr.FormattedAddr)
		//prints latitude and longitude for the given location
		fmt.Printf("lat:%f lng:%f\n", addr.Geometry.Location.Lat, addr.Geometry.Location.Lng)
	}
}
```

Another more sophisticated example shows the rate limiting feature of geopard
```Go
package main

import (
	"fmt"
	"sync"

	"github.com/dbriemann/geopard"
)

var (
	geocoder = geopard.GetInstance()

	cities = []string{
		"New York", "Los Angeles", "Chicago ", "Houston ", "Phoenix ", "Philadelphia ", "San Antonio ", "San Diego ", "Dallas ",
		"San Jose ", "Detroit ", "Jacksonville ", "Indianapolis", "San Francisco ", "Columbus ", "Austin ", "Memphis ", "Fort Worth ",
		"Baltimore ", "Charlotte ", "El Paso ", "Boston ", "Seattle ", "Washington ", "Milwaukee ", "Denver ", "Louisville/Jefferson County ",
		"Las Vegas ", "Nashville-Davidson ", "Oklahoma City ", "Portland ", "Tucson ", "Albuquerque ", "Atlanta ", "Long Beach ", "Fresno ",
		"Sacramento ", "Mesa ", "Kansas City ", "Cleveland ", "Virginia Beach ", "Omaha ", "Miami ", "Oakland ", "Tulsa ", "Honolulu ",
		"Minneapolis ", "Colorado Springs ", "Arlington ", "Wichita ", "Raleigh ", "St. Louis ", "Santa Ana ", "Anaheim ", "Tampa ",
		"Cincinnati ", "Pittsburgh ", "Bakersfield ", "Aurora ", "Toledo ", "Riverside ", "Stockton ", "Corpus Christi ", "Newark ",
		"Anchorage", "Buffalo ", "St. Paul ", "Lexington-Fayette", "Plano ", "Fort Wayne ", "St. Petersburg ", "Glendale ", "Jersey City ",
		"Lincoln ", "Henderson ", "Chandler ", "Greensboro ", "Scottsdale ", "Baton Rouge ", "Birmingham ", "Norfolk ", "Madison ",
		"New Orleans ", "Chesapeake ", "Orlando ", "Garland ", "Hialeah ", "Laredo ", "Chula Vista ", "Lubbock ", "Reno ", "Akron ",
		"Durham ", "Rochester ", "Modesto ", "Montgomery ", "Fremont ", "Shreveport ", "Arlington", "Glendale",
	}
)

func getLoc(name string, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	if loc, err := geocoder.Geocode(name); err != nil {
		fmt.Println(err.Error())
	} else {
		addr := loc.Results[0]
		//prints the formatted address for the queried location
		fmt.Println(addr.FormattedAddr)
		//prints latitude and longitude for the given location
		fmt.Printf("lat:%f lng:%f\n", addr.Geometry.Location.Lat, addr.Geometry.Location.Lng)
	}
}

func main() {
	defer geocoder.Destroy()

	//using a wait group to avoid premature termination of main
	var wg sync.WaitGroup
	wg.Add(len(cities))
	for _, c := range cities {
		//launching a goroutine for every location lookup
		//you will see the results printed in bursts of 10 (the limit per second of the google service)
		go getLoc(c, &wg)
	}

	//wait for all goroutines to finish
	wg.Wait()
}
```
