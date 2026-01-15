package main

import (
	"go-dns-server/internal/dns"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	addr := ":8083"
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s:%v", addr, err)
	}
	defer pc.Close()
	dns.LoadRecords("config.json")
	log.Printf("DNS Server is running on %s..", addr)
	go func() {
		for {
			buf := make([]byte, 512)
			n, clientAddr, err := pc.ReadFrom(buf)
			if err != nil {
				log.Printf("Read error: %v", err)
				continue
			}
			go dns.HandlePacket(pc, clientAddr, buf[:n])
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("\nShutting down DNS server...")
}
