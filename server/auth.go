package server

import (
	"crypto/sha1"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

// sharedKey gets the hash of the api key or passphrase that's configured in our server.
// If no key or passphrase is given, an array of 16 random bytes is returned.
func (s *Server) sharedKey() []byte {
	var b []byte
	switch {
	case s.Params.APIKey != nil:
		b = s.Params.APIKey
	case s.Params.Passphrase != "":
		hash := sha1.Sum([]byte(s.Params.Passphrase))
		b = hash[0:16]
		b = []byte(fmt.Sprintf("%x", b))
	default:
		b = make([]byte, 16)
		for i := range b {
			b[i] = byte(rand.Intn(256))
		}
		b = []byte(fmt.Sprintf("%x", b))
	}
	return b
}

func (s *Server) checkAPIKey(inputKey string) bool {
	if inputKey == "" {
		return false
	}
	k := s.sharedKey()
	keyToMatch := fmt.Sprintf("%s", k)
	if inputKey == keyToMatch {
		return true
	}
	log.Printf("Incorrect api key, expected %x", k)
	return false
}
