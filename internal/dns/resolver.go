package dns

import (
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"log"
	"net"
	"time"
)

const RootServer = "198.41.0.4"

func Resolve(question dnsmessage.Question) (*dnsmessage.Message, error) {
	currentServer := RootServer

	for i := 0; i < 7; i++ {
	nextHop:
		log.Printf("NAVIGATOR: [Hop %d] Asking %s for %s", i, currentServer, question.Name.String())

		reply, err := queryServer(currentServer, question)
		if err != nil {
			return nil, err
		}
		if len(reply.Answers) > 0 {
			log.Printf("NAVIGATOR: Success! Found IP for %s", question.Name.String())
			return reply, nil
		}
		if nextIP, found := findNextIP(reply); found {
			currentServer = nextIP
			continue
		}
		if len(reply.Authorities) > 0 {
			for _, auth := range reply.Authorities {
				if ns, ok := auth.Body.(*dnsmessage.NSResource); ok {
					nsName := ns.NS.String()
					nsReply, err := Resolve(dnsmessage.Question{
						Name:  dnsmessage.MustNewName(nsName),
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
					})
					if err == nil && len(nsReply.Answers) > 0 {
						if a, ok := nsReply.Answers[0].Body.(*dnsmessage.AResource); ok {
							currentServer = fmt.Sprintf("%d.%d.%d.%d", a.A[0], a.A[1], a.A[2], a.A[3])
							goto nextHop
						}
					}
				}
			}
		}

		return nil, fmt.Errorf("resolution failed: reached a dead end")
	}
	return nil, fmt.Errorf("resolution failed: too many hops")
}
func queryServer(server string, question dnsmessage.Question) (*dnsmessage.Message, error) {
	conn, err := net.DialTimeout("udp", net.JoinHostPort(server, "53"), 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:               0,
			RecursionDesired: false,
		},
		Questions: []dnsmessage.Question{question},
	}

	packed, _ := msg.Pack()
	_, err = conn.Write(packed)
	if err != nil {
		return nil, err
	}

	replyBuf := make([]byte, 1024)
	n, err := conn.Read(replyBuf)
	if err != nil {
		return nil, err
	}

	var reply dnsmessage.Message
	if err := reply.Unpack(replyBuf[:n]); err != nil {
		return nil, err
	}
	return &reply, nil
}
func findNextIP(msg *dnsmessage.Message) (string, bool) {
	for _, extra := range msg.Additionals {
		if extra.Header.Type == dnsmessage.TypeA {
			if a, ok := extra.Body.(*dnsmessage.AResource); ok {
				ip := fmt.Sprintf("%d.%d.%d.%d", a.A[0], a.A[1], a.A[2], a.A[3])
				return ip, true
			}
		}
	}
	return "", false
}
