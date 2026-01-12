package dns

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	records   = make(map[string]string)
	recordsMu sync.RWMutex
)

func LoadRecords(filename string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Could not read config file: %v", err)
		return
	}

	// We update the existing records map
	newRecords := make(map[string]string)
	err = json.Unmarshal(file, &newRecords)
	if err != nil {
		log.Printf("Could not parse JSON: %v", err)
		return
	}
	recordsMu.Lock()
	records = newRecords
	recordsMu.Unlock()
	log.Printf("Loaded %d local records from %s", len(records), filename)
}
func HandlePacket(pc net.PacketConn, addr net.Addr, buf []byte) {
	startTime := time.Now()
	fmt.Printf("DEBUG: Received packet from %s\n", addr.String())
	var m dnsmessage.Message
	if err := m.Unpack(buf); err != nil {
		log.Printf("Failed to unpack: %v", err)
		return
	}

	if len(m.Questions) == 0 {
		log.Printf("Error: No questions in packet")
		return
	}

	question := m.Questions[0]
	name := strings.ToLower(question.Name.String())
	log.Printf("2. Query received for domain: %s", name)

	if cachedMsg, found := GetCache(name); found && cachedMsg != nil {
		cachedMsg.Header.ID = m.Header.ID // Sync the ID
		packed, err := cachedMsg.Pack()
		if err == nil {
			pc.WriteTo(packed, addr)
			log.Printf("CACHE HIT: %s [Time: %v]", name, time.Since(startTime))
			return
		}
	}
	recordsMu.RLock()
	ip, ok := records[name]
	recordsMu.RUnlock()

	// Check if we have a local record
	if ok && question.Type == dnsmessage.TypeA {
		log.Printf("3a. Local record found for %s -> %s", name, ip)
		sendLocalResponse(pc, addr, m, ip)
		return
	}

	// If not local, use the Resolver (Step 2)
	log.Printf("3b. Not local. Forwarding %s to Google (8.8.8.8)...", name)
	resolved, err := Resolve(question)
	if err != nil && resolved == nil {
		log.Printf("Resolution error: %v", err)
		return
	}

	// Cache the resolved response
	SetCache(name, *resolved)

	// Sync the ID from the original query to the resolved response
	log.Printf("4. Google replied. Sending answer back to %s", addr)
	for _, answer := range resolved.Answers {
		if rec, ok := answer.Body.(*dnsmessage.AResource); ok {
			log.Printf(">> FOUND IP: %d.%d.%d.%d", rec.A[0], rec.A[1], rec.A[2], rec.A[3])
		}
	}
	resolved.Header.ID = m.Header.ID
	packed, _ := resolved.Pack()
	if err == nil {
		pc.WriteTo(packed, addr)
		log.Printf("CACHE MISS (Resolved): %s [Time: %v]", name, time.Since(startTime))
	}
}

func sendLocalResponse(pc net.PacketConn, addr net.Addr, request dnsmessage.Message, ip string) {
	// Parse string IP to [4]byte
	parsedIP := net.ParseIP(ip).To4()
	var ipBytes [4]byte
	copy(ipBytes[:], parsedIP)

	response := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID:                 request.Header.ID,
			Response:           true,
			Authoritative:      true,
			RecursionAvailable: true,
		},
		Questions: request.Questions,
		Answers: []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{
					Name:  request.Questions[0].Name,
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
					TTL:   600,
				},
				Body: &dnsmessage.AResource{A: ipBytes},
			},
		},
	}

	packed, _ := response.Pack()
	pc.WriteTo(packed, addr)
}
