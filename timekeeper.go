package main

// Description: This file contains the TimeKeeper function, which is used to keep track of time, and perform actions based on time.

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

// TimeKeeper i used to trigger timed events.
func TimeKeeper() {
	System.WriteToChannel(msg{payload: []byte("TimeKeeper: Starting TimeKeeper")})
	var catMessages []string
	SpecialMessage = "Awaiting new Cat Message within 5 minutes."
	fivesecond := time.NewTicker(5 * time.Second) // five second ticker
	fiveminute := time.NewTicker(5 * time.Minute) // five minute ticker
	defer func() {
		System.LogPayload(timekeeper, "Closing TimeKeeper")
		fivesecond.Stop()
		fiveminute.Stop()
	}()
	for {
		select {
		// We ingest the file ever five seconds, and populat the catMessages slice.
		// This way we can live update the banner message file.
		case <-fivesecond.C:
			bytes, err := ioutil.ReadFile(banmsgfp) // read banner message file
			if err != nil {
				Error.LogError(timekeeper, fmt.Sprintf("Read failed(%v): %v", banmsgfp, err.Error()))
			}
			catMessages = strings.Split(string(bytes), "\n") // split banner message file into slice
		case <-fiveminute.C:
			System.WriteToChannel(msg{payload: []byte("TimeKeeper: 5 minutes have passed. Setting a new message")})
			SpecialMessage = catMessages[rand.Intn(len(catMessages))] // choose random message from slice
			System.WriteToChannel(msg{payload: []byte("Choosing new cat message. '" + SpecialMessage + "")})

		}
	}
}
