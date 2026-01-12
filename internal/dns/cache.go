package dns

import (
	"log"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

type cacheEntry struct {
	msg     dnsmessage.Message
	expires time.Time
}

var (
	dnsCache sync.Map
)

func GetCache(name string) (*dnsmessage.Message, bool) {
	val, ok := dnsCache.Load(name)
	if !ok {
		return nil, false
	}

	entry := val.(cacheEntry)
	if time.Now().After(entry.expires) {
		dnsCache.Delete(name)
		return nil, false
	}
	return &entry.msg, true
}

func SetCache(name string, msg dnsmessage.Message) {
	var ttl uint32 = 60
	if len(msg.Answers) > 0 {
		ttl = msg.Answers[0].Header.TTL
	}
	dnsCache.Store(name, cacheEntry{
		msg:     msg,
		expires: time.Now().Add(time.Duration(ttl) * time.Second),
	})
	log.Printf("Caching %s for %d seconds", name, ttl)
}
