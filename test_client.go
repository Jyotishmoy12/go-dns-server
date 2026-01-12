package main

import (
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"time"
)

func main() {
	targetDomain := "amazon.com."

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{ID: 1234, RecursionDesired: true},
		Questions: []dnsmessage.Question{
			{
				Name:  dnsmessage.MustNewName(targetDomain),
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
			},
		},
	}
	packed, _ := msg.Pack()

	conn, _ := net.Dial("udp", "127.0.0.1:8083")
	defer conn.Close()

	fmt.Printf("Sending query for %s to server...\n", targetDomain)
	conn.Write(packed)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	reply := make([]byte, 512)
	n, err := conn.Read(reply)
	if err != nil {
		fmt.Printf("Error: No response from server: %v\n", err)
		return
	}

	var res dnsmessage.Message
	if err := res.Unpack(reply[:n]); err != nil {
		fmt.Printf("Error unpacking: %v\n", err)
		return
	}
	if len(res.Answers) > 0 {
		fmt.Printf("\n--- DNS SERVER RESULTS ---\n")
		for i, answer := range res.Answers {
			if a, ok := answer.Body.(*dnsmessage.AResource); ok {
				fmt.Printf("[%d] IP Address: %d.%d.%d.%d (TTL: %d)\n",
					i+1, a.A[0], a.A[1], a.A[2], a.A[3], answer.Header.TTL)
			}
		}
	} else {
		fmt.Println("Server replied but there were no answers.")
	}
}
