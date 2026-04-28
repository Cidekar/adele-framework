package urlsigner

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	// create url and make sure it expires at a given time
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	// Append a hash to the url that our system knows about
	// does the url have querystring parameters in the url
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}

	tokenBytes := crypt.Sign([]byte(urlToSign))
	token := string(tokenBytes)

	return token
}

func (s *Signer) VerifyToken(token string) bool {
	// create url and make sure it expires at a given time
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))
	if err != nil {
		return false
	}
	return true
}

func (s *Signer) Expired(token string, minutesUntilExpired int) bool {
	// create url and make sure it expires at a given time
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	timestamp := crypt.Parse([]byte(token))

	return time.Since(timestamp.Timestamp) > time.Duration(minutesUntilExpired)*time.Minute
}
