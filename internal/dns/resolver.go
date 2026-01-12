package dns

import (
	"log"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const UpstreamDNS = "8.8.8.8:53"

// Resolve forwards a query to an upstream provider
func Resolve(question dnsmessage.Question) (*dnsmessage.Message, error) {
	// pack the question into a dns message
	msg := dnsmessage.Message{
		Header:    dnsmessage.Header{ID: 0, RecursionDesired: true},
		Questions: []dnsmessage.Question{question},
	}
	packed, err := msg.Pack()
	if err != nil {
		return nil, err
	}
	// open udp connection to upstream (Google)
	log.Printf("RESOLVER: Dialing Google...")
	conn, err := net.DialTimeout("udp", UpstreamDNS, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// send and receives
	log.Printf("RESOLVER: Writing packet to Google...")
	_, err = conn.Write(packed)
	if err != nil {
		return nil, err
	}
	log.Printf("RESOLVER: Waiting for Google's reply...")
	reply := make([]byte, 512)
	n, err := conn.Read(reply)
	if err != nil {
		return nil, err
	}
	log.Printf("RESOLVER: Google replied with %d bytes!", n)
	var response dnsmessage.Message
	err = response.Unpack(reply[:n])
	if err != nil {
		return nil, err
	}
	return &response, nil
}
