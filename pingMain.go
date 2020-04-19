// This file contains the executable main function.
// This package contains a CLI app which accepts a hostname or an IP address as its argument,
// then sends ICMP "echo requests" in a loop to the target while receiving "echo reply" messages.
// It also report loss and RTT times for each sent message.
//
// It also gives users features like:
// 1. Set limit: Sets the limit to maximum successful requests. (Use -l=<limit>)
// 2. Set interval: Allows user to set how often a packet is sent. (Use -i=<interval>)
// 3. Display stats: Displays various statistics about packets and RTT (Use -stats=true)
// 4. RTT limit: In case this limit is surpassed report the user. (Use -rtt=<limit>)
//
//
// Author: Aashish Thakur
package main

import (
	"flag"
	"fmt"
	"time"
)

const (
	message = string("Allows the user to select the maximum amount of the pings")
	rttMessage = string("Allows the user to select an RTT limit in ms")
	statsMessage = string("Allows the user to select the display of stats")
	intervalMessage = string("Allows the user to select interval after which every packet is sent")
)

func main(){

	limit := flag.Int("l",0,message)
	rttLimit := flag.Int64("rtt",0,rttMessage)
	getStats := flag.Bool("stats",false,statsMessage)
	interval := flag.Duration("i", time.Second, intervalMessage)

	flag.Parse() //Get host name/link

	if flag.NArg()==0 {
		flag.Usage()
		return
	}

	hostname := flag.Arg(0)
	pingHandler,err := ping(hostname)
	if err != nil{
		fmt.Print("Encountered Error: ", err.Error())
		return
	}

	pingHandler.rttLimit = *rttLimit
	pingHandler.limit = *limit
	pingHandler.interval = *interval
	pingHandler.getStats = *getStats

	fmt.Printf("Source: %s (%s) \n", hostname, pingHandler.ip)
	pingHandler.start()

}

