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
	// run holds the setup and teardown in a function so deferred calls
	// execute before os.Exit, which bypasses them.
	os.Exit(run(m))
}

func run(m *testing.M) int {
	// Bound setup so a cold image pull cannot hang forever.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start MailHog container using the generic container API.
	mailhogContainer, err := testcontainers.Run(ctx,
		"mailhog/mailhog:latest",
		testcontainers.WithExposedPorts("1025/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("1025/tcp").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Printf("could not start MailHog container: %v", err)
		return 1
	}
	// Terminate once the container has started, on every exit path.
	defer func() {
		termCtx, termCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer termCancel()
		if err := mailhogContainer.Terminate(termCtx); err != nil {
			log.Printf("could not terminate MailHog container: %v", err)
		}
	}()

	// Point the mailer at the real mapped SMTP host and port.
	host, err := mailhogContainer.Host(ctx)
	if err != nil {
		log.Printf("could not get container host: %v", err)
		return 1
	}

	port, err := mailhogContainer.MappedPort(ctx, "1025/tcp")
	if err != nil {
		log.Printf("could not get mapped SMTP port: %v", err)
		return 1
	}

	mailer.Host = host
	mailer.Port = int(port.Num())

	go mailer.ListenForMail()

	return m.Run()
}
