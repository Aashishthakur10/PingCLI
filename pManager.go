// The executable file in pingMain.go
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
	"encoding/binary"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"math/rand"
	"net"
	"time"
)

const (
	ipv4Len    = 4
	ListenAddr = ""
)

//Structure for the ping handle object
type pingHandler struct {
	limit     int           // In case user sets a limit
	interval  time.Duration // Period after which ICMP message is sent
	rtt       time.Duration // Amount of round-trip time taken for a packet
	ip        *net.IPAddr   // IP address
	isIPv4    bool          // check if it's ipv4 or ipv6
	seqNo     int           // Sequence number
	ID        int           // Identifier
	totalSent int           // Total packet sent, used for stats
	totalRecv int           // Total packet Received, used for stats
	rttLimit  int64         // rtt limit set by user, flag if it's crossed.
	getStats  bool          // Display stats
}

//Structure to display stats
type stats struct {
	minRTT time.Duration
	maxRTT time.Duration
	total  time.Duration
}

// Structure to store incoming packet response
type recvPacket struct {
	rtt   time.Duration
	bytes int
	seqNo int
	ip    *net.IPAddr
	ttl   int
	data  []byte
}

// Main function
func ping(ip string) (*pingHandler, error) {
	ipStruct, err := net.ResolveIPAddr("", ip)
	if err != nil {
		return nil, err
	}

	return &pingHandler{
		ip:        ipStruct,
		isIPv4:    isIPv4(ipStruct.IP),
		ID:        rand.Intn(1000000),
		totalSent: 0,
	}, err
}

// Check if ipv4 or ipv6.
func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == ipv4Len
}

// Send, receives and processes ICMP packets and also multiple user inputs.
// IT's handled for both ipv4 and ipv6.
//
// The user has the following options:
// 1. Set limit: Sets the limit to maximum successful requests.
// 2. Set interval: Allows user to set how often a packet is sent.
// 3. Display stats: Displays various statistics about packets and RTT
// 4. RTT limit: In case this limit is surpassed report the user.

func (ping *pingHandler) start() {

	var ipType string

	//Check if the ip type is for ipv4/ipv6
	if ping.isIPv4 {
		ipType = "udp4"
	} else {
		ipType = "udp6"
	}

	c, err := icmp.ListenPacket(ipType, ListenAddr)
	c.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)

	defer c.Close()
	if err != nil {
		fmt.Println("Error in listening: ", err.Error())
		c.Close()
		return
	}

	statBot := stats{
		minRTT: time.Duration(10000000000000),
		maxRTT: 0,
		total:  0,
	}
	//Interval after which the packets will be sent.
	interval := time.NewTicker(ping.interval)
	defer interval.Stop()

	for {
		select {
		case <-interval.C:
			receivedAt := time.Now()
			// Send the ICMP packet to the host.
			err := ping.sendICMP(c)
			if err != nil {
				fmt.Println(err.Error())
			}
			// Process the packet by first listening to the response.
			preprossedPacket, err := ping.receiveICMP(c)

			if err != nil {
				continue
			}
			// Now break down the message received.
			err = ping.processPacket(preprossedPacket)

			if err != nil {
				fmt.Print("An error has occured: ", err.Error())
			}

			preprossedPacket.rtt = time.Since(receivedAt)

			// Calculations for min and max RTT
			if preprossedPacket.rtt < statBot.minRTT {
				statBot.minRTT = preprossedPacket.rtt
			}
			if preprossedPacket.rtt > statBot.maxRTT {
				statBot.maxRTT = preprossedPacket.rtt
			}
			statBot.total += preprossedPacket.rtt

			ping.handleDisplay(preprossedPacket, &statBot)

			if ping.limit > 0 && ping.totalRecv >= ping.limit {
				c.Close()
				interval.Stop()
				return
			}
		}
	}
}

// Display various information such as packet details, latency and stats.
func (ping *pingHandler) handleDisplay(preprossedPacket *recvPacket,
	statBot *stats) {

	loss := float64(ping.totalSent-ping.totalRecv) / float64(ping.totalSent) * 100

	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v packet loss = %v\n",
		preprossedPacket.bytes, preprossedPacket.ip, preprossedPacket.seqNo,
		preprossedPacket.rtt, preprossedPacket.ttl, loss)

	if ping.rttLimit > 0 && preprossedPacket.rtt.Milliseconds() > ping.rttLimit {
		fmt.Printf("Crossed the user set RTT limit with RTT: %v \n", preprossedPacket.rtt)
	}
	if ping.getStats {
		fmt.Print("Stats: \n")
		avgRTT := statBot.total / time.Duration(ping.totalRecv)
		fmt.Printf("Total sent=%v total received=%v RTT-min=%v RTT-avg=%v RTT-max=%v\n",
			ping.totalSent, ping.totalRecv, statBot.minRTT, avgRTT, statBot.maxRTT)
	}
}

// Process the packet by breaking down the message
// received from server.
func (ping *pingHandler) processPacket(packet *recvPacket) error {
	var proto int

	if ping.isIPv4 {
		proto = 1
	} else {
		proto = 58
	}

	msg, err := icmp.ParseMessage(proto, packet.data)

	if err != nil {
		return err
	}
	// Only concerned with message type echo reply.
	if msg.Type != ipv4.ICMPTypeEchoReply && msg.Type != ipv6.ICMPTypeEchoReply {
		return nil
	}
	_, ok := msg.Body.(*icmp.Echo)
	if ok {
		ping.totalRecv++
	}
	return nil

}

// Handle sending of an ICMP packet to host, involves
// using multiple structures to properly format data
// in body and message.
//
// In case the packets sending is not successful,
// resend the packet. Once successful return.
func (ping *pingHandler) sendICMP(c *icmp.PacketConn) error {

	destination := &net.UDPAddr{IP: ping.ip.IP, Zone: ping.ip.Zone}
	var requestType icmp.Type
	// Echo type for ipv4 and ipv6
	if ping.isIPv4 {
		requestType = ipv4.ICMPTypeEcho
	} else {
		requestType = ipv6.ICMPTypeEchoRequest
	}
	// Data being sent.
	t := intToBytes(time.Now().Unix())
	body := &icmp.Echo{
		ID:   ping.ID,
		Seq:  ping.seqNo,
		Data: t,
	}

	message := &icmp.Message{
		Type: requestType,
		Code: 0,
		Body: body,
	}

	msg, err := message.Marshal(nil)

	if err != nil {
		return err
	}

	for {
		_, err = c.WriteTo(msg, destination)
		ping.totalSent++
		if err == nil {
			ping.seqNo++
			break
		}
		duration := time.Second
		time.Sleep(duration)
	}

	return nil
}

// Handle response received from the host.
func (ping *pingHandler) receiveICMP(c *icmp.PacketConn) (*recvPacket, error) {

	reply := make([]byte, 512)
	var TTL int
	var n int
	var err error
	c.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	// Different control messages for ipv4 and ipv6.
	if ping.isIPv4 {
		var cm *ipv4.ControlMessage
		n, cm, _, err = c.IPv4PacketConn().ReadFrom(reply)
		if cm != nil {
			TTL = cm.TTL
		}
	} else {
		var cm *ipv6.ControlMessage
		n, cm, _, err = c.IPv6PacketConn().ReadFrom(reply)
		if cm != nil {
			TTL = cm.HopLimit
		}
	}

	if err != nil {
		return nil, err
	}

	return &recvPacket{
		bytes: n,
		seqNo: ping.seqNo,
		ip:    ping.ip,
		ttl:   TTL,
		data:  reply,
	}, nil
}

// Covert int to bytes.
func intToBytes(tracker int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(tracker))
	return b
}
