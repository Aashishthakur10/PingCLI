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
	ipv4Len = 4
	ListenAddr = ""
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
	totalRecv int // Total packet Recieved, used for stats

}

type recvPacket struct {
	rtt time.Duration
	bytes int
	seqNo int
	ip *net.IPAddr
	ttl int
	data []byte
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

	var ipType string

	if ping.isIPv4{
		ipType = "udp4"
	}else {
		ipType = "udp6"
	}

	//c,err := icmp.ListenPacket(ipType,ListenAddr)
	c, err := icmp.ListenPacket(ipType, ListenAddr)
	c.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)

	defer c.Close()
	if  err!=nil{
		fmt.Println("Error in listening: ", err.Error())
		c.Close()
		return
	}

	for  {
		receivedAt := time.Now()
		err := ping.sendICMP(c)
		if err != nil {
			fmt.Println(err.Error())
		}
		preprossedPacket,err :=ping.receiveICMP(c)
		if err!=nil {
			fmt.Printf("Error is: %s", err.Error())
		}

		err = ping.processPacket(preprossedPacket)
		preprossedPacket.rtt = time.Now().Sub(receivedAt)
		if err !=nil{
			fmt.Errorf(err.Error())
		}
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
			preprossedPacket.bytes, preprossedPacket.ip, preprossedPacket.seqNo,
			preprossedPacket.rtt, preprossedPacket.ttl)
		if ping.limit>0 && ping.totalRecv >= ping.limit {
			c.Close()
			return
		}
		duration := time.Second
		time.Sleep(duration)
	}
}


func  (ping *pingHandler)processPacket(packet *recvPacket) error {
	var proto int
	if  ping.isIPv4{
		proto = 1
	}else {
		proto = 58
	}
	msg,err:= icmp.ParseMessage(proto,packet.data)

	if err!=nil {
		return err
	}
	if msg.Type != ipv4.ICMPTypeEchoReply && msg.Type != ipv6.ICMPTypeEchoReply {
		return nil
	}
	 _, ok := msg.Body.(*icmp.Echo)
	if ok {
		ping.totalRecv++

	}
	return nil

}

func (ping *pingHandler)sendICMP(c *icmp.PacketConn) error{
	destination := &net.UDPAddr{IP: ping.ip.IP, Zone: ping.ip.Zone}
	var requestType icmp.Type
	if ping.isIPv4 {
		requestType = ipv4.ICMPTypeEcho
	} else {
		requestType = ipv6.ICMPTypeEchoRequest
	}

	t := append(timeToBytes(time.Now()), intToBytes(1839682523836159141)...)
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


func (ping *pingHandler)receiveICMP(c *icmp.PacketConn, ) (*recvPacket,error) {
	//for {
		reply := make([]byte, 512)
		var TTL int
		var n int
		var err error
		c.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		if ping.isIPv4 {
			var cm *ipv4.ControlMessage
			n, cm, _, err = c.IPv4PacketConn().ReadFrom(reply)
			if cm!=nil {
				TTL = cm.TTL
			}
		}else {
			var cm *ipv6.ControlMessage
			n, cm, _, err = c.IPv6PacketConn().ReadFrom(reply)
			if cm!=nil {
				TTL = cm.HopLimit
			}
		}

	if err!=nil {
		return nil,err
	}

	return &recvPacket{
		bytes: n,
		seqNo: ping.seqNo,
		ip:    ping.ip,
		ttl:   TTL,
		data: reply,
	},nil
}



func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}


func intToBytes(tracker int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(tracker))
	return b
}
