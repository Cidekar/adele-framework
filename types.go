package adele

import (
	"net"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/cidekar/adele-framework/cache"
	"github.com/cidekar/adele-framework/database"
	"github.com/cidekar/adele-framework/helpers"
	"github.com/cidekar/adele-framework/mailer"
	"github.com/cidekar/adele-framework/middleware"
	"github.com/cidekar/adele-framework/mux"
	"github.com/cidekar/adele-framework/provider"
	"github.com/cidekar/adele-framework/render"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Adele struct {
	AppName          string
	config           config
	Cache            cache.Cache
	DB               *database.Database
	Debug            bool
	FileSystem       map[string]interface{}
	Helpers          *helpers.Helpers
	JetViews         *jet.Set
	Log              *logrus.Logger
	Mail             mailer.Mail
	middleware       middleware.Middleware
	MaintenanceMode  bool
	Provider         *provider.Provider
	Render           *render.Render
	Routes           *mux.Mux
	RootPath         string
	RPCListener      *net.Listener
	Scheduler        *cron.Cron
	Session          *scs.SessionManager
	Version          string
	ViewsTemplateDir string
}

type config struct {
	port        string
	renderer    string
	sessionType string
}
