package mailer

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var mailer = Mail{
	Domain:      "localhost",
	Templates:   "./testdata/mail",
	Host:        "localhost",
	Encryption:  "none",
	FromAddress: "me@here.com",
	FromName:    "Adel",
	Jobs:        make(chan Message, 1),
	Results:     make(chan Result, 1),
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start MailHog container using the generic container API.
	mailhogContainer, err := testcontainers.Run(ctx,
		"mailhog/mailhog:latest",
		testcontainers.WithExposedPorts("1025/tcp", "8025/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("1025/tcp").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("could not start MailHog container: %v", err)
	}

	// Point the mailer at the real mapped SMTP host and port.
	host, err := mailhogContainer.Host(ctx)
	if err != nil {
		log.Fatalf("could not get container host: %v", err)
	}

	port, err := mailhogContainer.MappedPort(ctx, "1025/tcp")
	if err != nil {
		log.Fatalf("could not get mapped SMTP port: %v", err)
	}

	mailer.Host = host
	mailer.Port = port.Int()

	go mailer.ListenForMail()

	code := m.Run()

	// Terminate explicitly before os.Exit, which bypasses deferred calls.
	if err := mailhogContainer.Terminate(ctx); err != nil {
		log.Printf("could not terminate MailHog container: %v", err)
	}

	os.Exit(code)
}
