//go:build aerra_template

package main

import (
	"log"
	"$APPNAME$/handlers"
	"$APPNAME$/middleware"
	"$APPNAME$/models"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cidekar/adele-framework"
	"github.com/cidekar/adele-framework/httpserver"
	"github.com/cidekar/adele-framework/provider"
	"github.com/cidekar/adele-framework/rpcserver"
)

var wg sync.WaitGroup

func main() {

	a := bootstrapApplication()

	go a.Mail.ListenForMail()

	go a.listenForShutdown()

	err := rpcserver.Start(a.App)
	if err != nil {
		log.Fatalf("failed to start rpc: %s", err)
	}

	a.jobsSchedule()

	err = httpserver.Start(a.App)

	a.App.Log.Error(err)

}

func (a *application) listenForShutdown() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	a.App.Log.Info("Application received signal", s.String())

	err := rpcserver.Stop(a.App)
	if err != nil {
		log.Fatal("RPC server failed to stop:", err)
	}

	a.App.Log.Info("Good bye!")

	os.Exit(0)
}

func (a *application) jobsSchedule() {
	// ...
}

func bootstrapApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	a := &adele.Adele{}
	err = a.New(path)
	if err != nil {
		log.Fatal(err)
	}

	a.AppName = "$APPNAME$"

	// Construct Models first so Handlers and Middleware can be wired against
	// the same instance — h.Models / m.Models would otherwise be nil and any
	// receiver that touches them (e.g. Registration, Login, CheckRemember)
	// would nil-panic at request time.
	m := models.New(a)

	myMiddleware := &middleware.Middleware{
		App:    a,
		Models: m,
	}

	myHandlers := &handlers.Handlers{
		App:    a,
		Models: m,
	}

	app := &application{
		App:        a,
		Handlers:   myHandlers,
		Mail:       &a.Mail,
		Middleware: myMiddleware,
		Models:     m,
	}

	app.App.Routes = app.routes()

	p := &provider.Provider{
		EnabledProviders: make(map[string]bool),
		ProviderConfigs:  make(map[string]map[string]interface{}),
	}

	a.Provider = p

	if err := a.Provider.LoadProviders(app.App); err != nil {
		a.Log.Error(err)
		os.Exit(1)
	}

	return app
}
