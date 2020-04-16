package Main

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"math/rand"
	"net"
	"time"
)

const (
	ipv4Len = 4
	ListenAddr = "0.0.0.0"
)

//Structure for the ping handle object
type pingHandler struct {

	limit int		// In case user sets a limit
	rtt time.Duration // Amount of round-trip time taken for a packet
	ip *net.IPAddr // IP address
	isIPv4 bool // check if it's ipv4 or ipv6
	seqNo int 	// Sequence number
	ID   int    // identifier
	totalSent int // Total packet sent, used for stats

}

type recvPacket struct {

}

func ping(ip string, limit int) (*pingHandler,error) {
	ipStruct, err := net.ResolveIPAddr("", ip)
	if err != nil{
		return nil,err
	}

	return &pingHandler{
		limit:  limit,
		ip:     ipStruct,
		isIPv4: isIPv4(ipStruct.IP),
		ID:     rand.Intn(1000000),
		totalSent: 0,
	},err
}

// Check if ipv4 or ipv6.
func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == ipv4Len
}

func (ping *pingHandler)start(){

	var ipType = ""

	if ping.isIPv4{
		ipType = "ip4:icmp"
	}else {
		ipType = "ip6:ipv6-icmp"
	}

	c,err := icmp.ListenPacket(ipType,ListenAddr)

	defer c.Close()
	if  err!=nil{
		fmt.Println("Error in listening: ", err.Error())
		c.Close()
		return
	}
}



func (ping *pingHandler)sendICMP(c *icmp.PacketConn) error{
	destination := &net.UDPAddr{IP: ping.ip.IP, Zone: ping.ip.Zone}
	var requestType icmp.Type
	if ping.isIPv4 {
		requestType = ipv4.ICMPTypeEcho
	} else {
		requestType = ipv6.ICMPTypeEchoRequest
	}

	body := &icmp.Echo{
		ID:   ping.ID,
		Seq:  ping.seqNo,
		Data: []byte(""),
	}

	message := &icmp.Message{
		Type: requestType,
		Code: 0,
		Body: body,
	}

	msg,err := message.Marshal(nil)

	if err!=nil {
		return err
	}


	for {
		_, err = c.WriteTo(msg, destination)
		if err == nil {
			ping.totalSent++
			ping.seqNo++
			break
		}
	}

	return nil
}




