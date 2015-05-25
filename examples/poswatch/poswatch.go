// poswatch.go
//
// Monitor PiBaHeAlTas and report its values once it starts udpating

package main

import (
	"fmt"
	"strings"
	"strconv"
	"time"
	"os"
	"math"
	"github.com/kuroneko/psx.go"
)

var (
	// we've received one update
	dataValid = false

	// data direct from PSX
	pitch		float64
	bank		float64
	heading		float64
	altitude	int64
	tas 		int64
	latitude	float64
	longitude	float64
)

// Receive an update for PiBaHeAlTas
func updatePosition(_ *psx.Connection, msg *psx.WireMsg) {
	// whilst psx.WireMsg may provide ValueAtSubIndex, it's more 
	// efficient to use Split if we're using all the values.
	msgParts := strings.Split(msg.Value, ";")

	pitch, _ = strconv.ParseFloat(msgParts[0], 64)
	bank, _ = strconv.ParseFloat(msgParts[1], 64)
	heading, _ = strconv.ParseFloat(msgParts[2], 64)
	altitude, _ = strconv.ParseInt(msgParts[3], 10, 64)
	tas, _ = strconv.ParseInt(msgParts[4], 10, 64)
	latitude, _ = strconv.ParseFloat(msgParts[5], 64)
	longitude, _ = strconv.ParseFloat(msgParts[6], 64)

	dataValid = true
}

func connectionLoop(pconn *psx.Connection) {
	for {
		err := pconn.Connect()
		if (err != nil) {
			fmt.Printf("Couldn't connect : %s\n", err)
			os.Exit(1)
		}
		pconn.Listener()
	}
}

func main() {
	pconn, err := psx.NewConnection("localhost:10747", "poswatch")
	if (err != nil) {
		fmt.Printf("Couldn't initialise connection: %s\n", err)
		os.Exit(1)
	}
	// connect up the callback
	pconn.Hooks["PiBaHeAlTas"] = updatePosition
	// if we're using SwitchPSX/Router, request only PiBaHeAlTas
	pconn.Subscribe("PiBaHeAlTas")

	go connectionLoop(pconn)

	for {
		if (dataValid) {
			// do some quick and dirty conversions...
			pitchDeg := pitch * 180.0 / math.Pi
			bankDeg := bank * 180.0 / math.Pi
			headingDeg := heading * 180.0 / math.Pi

			altitudeFmted := float64(altitude) / 1000.0
			tasFmted := float64(tas) / 1000.0

			latDeg := latitude * 180.0 / math.Pi
			longDeg := longitude * 180.0 / math.Pi

			fmt.Printf("Pitch: %.1f  Bank: %.1f  Heading: %.1f  Altitude: %.0f  TAS:  %.2f  Lat: %.4f  Long: %.4f\n",
				pitchDeg, bankDeg, headingDeg, altitudeFmted, tasFmted, latDeg, longDeg)
		}
		time.Sleep(time.Second)
	}
}
