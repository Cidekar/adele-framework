package urlsigner

import (
	"strings"
	"testing"
)

func TestUrlSigner_GenerateToken(t *testing.T) {

	sign := Signer{
		Secret: []byte("key"),
	}

	link := "https://localhost"
	signedLink := sign.GenerateTokenFromString(link)

	if strings.Contains(signedLink, "?hash=") == false {
		t.Error("url signer did not attach hash to link")
	}
}

func TestUrlSigner_VerifyToken(t *testing.T) {

	sign := Signer{
		Secret: []byte("key"),
	}

	link := "https://localhost"
	signedLink := sign.GenerateTokenFromString(link)

	if sign.VerifyToken(signedLink) == false {
		t.Error("url signer not able to verify token")
	}
}

func TestUrlSigner_Expired(t *testing.T) {

	sign := Signer{
		Secret: []byte("key"),
	}

	link := "https://localhost"
	signedLink := sign.GenerateTokenFromString(link)

	if sign.Expired(signedLink, 1) != false {
		t.Error("token expired when it should not be")
	}

	if sign.Expired(signedLink, 0) != true {
		t.Error("token is not expired when it be")
	}
}
