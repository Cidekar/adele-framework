package adele

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/cidekar/adele-framework/cache"
	"github.com/cidekar/adele-framework/cache/badgerdriver"
	"github.com/cidekar/adele-framework/cache/redisdriver"
	"github.com/cidekar/adele-framework/database"
	"github.com/cidekar/adele-framework/filesystem/miniofilesystem"
	"github.com/cidekar/adele-framework/filesystem/s3filesystem"
	"github.com/cidekar/adele-framework/filesystem/sftpfilesystem"
	"github.com/cidekar/adele-framework/filesystem/webdavfilesystem"
	"github.com/cidekar/adele-framework/helpers"
	"github.com/cidekar/adele-framework/logger"
	"github.com/cidekar/adele-framework/mailer"
	"github.com/cidekar/adele-framework/middleware"
	"github.com/cidekar/adele-framework/mux"
	"github.com/cidekar/adele-framework/render"
	"github.com/cidekar/adele-framework/session"
	crs "github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

const Version = "v1.0.2"

// Create a global helper instance for the package— provides access to all
// helper methods in sub-packages.
var Helpers = &helpers.Helpers{}

// Create a new instance of the Adele type using a pointer to Adele with the
// root path of the application as a argument. The new-up is called by project adele's consuming package
// to bootstrap the framework.
func (a *Adele) New(rootPath string) error {

	directories := []string{"handlers", "logs", "jobs", "middleware", "migrations", "models", "public", "resources", "resources/views", "resources/mail", "storage"}

	err := a.CreateDirectories(rootPath, directories)
	if err != nil {
		return err
	}

	err = a.CreateEnvironmentFile(rootPath)
	if err != nil {
		return err
	}

	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	a.Log = logger.CreateLogger()

	sess, err := a.BootstrapSessionManager()
	if err != nil {
		return err
	}

	a.Session = sess

	a.BootstrapMiddleware()

	a.Debug, _ = strconv.ParseBool(os.Getenv("APP_DEBUG"))
	a.RootPath = rootPath
	a.Version = Version
	a.ViewsTemplateDir = Helpers.Getenv("VIEWS_TEMPLATE_DIR", "resources/views")
	a.config = config{
		port:        Helpers.Getenv("HTTP_PORT", "4000"),
		renderer:    Helpers.Getenv("RENDERER", "jet"),
		sessionType: Helpers.Getenv("SESSION_TYPE"),
	}

	muxRouter, err := a.BootstrapMux(rootPath)
	if err != nil {
		return err
	}

	a.Routes = muxRouter.(*mux.Mux)

	a.BoostrapFilesystem()

	a.Mail = a.BoootstrapMailer()

	a.JetViews = a.BootstrapJetEngine()

	a.Render = a.BootstrapRender()

	a.BootstrapDatabase()

	a.Helpers = a.BootstrapHelpers()

	a.BootstrapScheduler()

	err = a.BootstrapCache(rootPath)
	if err != nil {
		return err
	}

	return nil
}

// Initializes and sets up a database connection for the application—establishes a database
// connection during application startup and stores it in the Adele struct.
func (a *Adele) BootstrapDatabase() {
	db, err := database.OpenDB(os.Getenv("DATABASE_TYPE"), &database.DataSourceName{
		Host:         Helpers.Getenv("DATABASE_HOST", "localhost"),
		Port:         Helpers.Getenv("DATABASE_PORT", "5432"),
		User:         Helpers.Getenv("DATABASE_USER"),
		Password:     Helpers.Getenv("DATABASE_PASSWORD"),
		DatabaseName: Helpers.Getenv("DATABASE_NAME"),
		SslMode:      Helpers.Getenv("DATABASE_SSL_MODE"),
	})

	if err != nil {
		a.Log.Error(err)
		os.Exit(1)
	}
	a.DB = &database.Database{
		DataType: os.Getenv("DATABASE_TYPE"),
		Pool:     db,
	}
}

// Initializes the file system auto-configuration method for the framework by detecting and
// initializes available file storage systems based on environment variables during application startup.
func (a *Adele) BoostrapFilesystem() {
	fileSystem := make(map[string]interface{})

	if os.Getenv("S3_KEY") != "" {
		s3 := s3filesystem.S3{
			Key:    os.Getenv("S3_KEY"),
			Secret: os.Getenv("S3_SECRET"),
			Region: os.Getenv("S3_REGION"),
			Bucket: os.Getenv("S3_BUCKET"),
		}
		fileSystem["S3"] = s3
	}

	if os.Getenv("MINIO_SECRET") != "" {
		useSSL := false
		if strings.ToLower(os.Getenv("MINIO_USESSL")) == "true" {
			useSSL = true
		}

		minio := miniofilesystem.Minio{
			Endpoint: os.Getenv("MINIO_ENDPOINT"),
			Key:      os.Getenv("MINIO_KEY"),
			Secret:   os.Getenv("MINIO_SECRET"),
			UseSSL:   useSSL,
			Region:   os.Getenv("MINIO_REGION"),
			Bucket:   os.Getenv("MINIO_BUCKET"),
		}
		fileSystem["MINIO"] = minio
	}

	if os.Getenv("SFTP_HOST") != "" {
		sftp := sftpfilesystem.SFTP{
			Host:     os.Getenv("SFTP_HOST"),
			User:     os.Getenv("SFTP_USER"),
			Password: os.Getenv("SFTP_PASSWORD"),
			Port:     os.Getenv("SFTP_PORT"),
		}
		fileSystem["SFTP"] = sftp
	}

	if os.Getenv("WEBDAV_HOST") != "" {
		webDAV := webdavfilesystem.WebDAV{
			Host:     os.Getenv("WEBDAV_HOST"),
			User:     os.Getenv("WEBDAV_USER"),
			Password: os.Getenv("WEBDAV_PASSWORD"),
		}
		fileSystem["WEBDAV"] = webDAV
	}

	a.FileSystem = fileSystem
}

// Creates and returns a helper utilities object for the Adele framework— a collection of utility functions
// that can be used throughout the application.
func (a *Adele) BootstrapHelpers() *helpers.Helpers {

	// Define the file types allowd by the system and add any provided by the application developer.
	mimeTypes := []string{"image/gif", "image/jpeg", "image/png", "application/pdf"}
	exploded := strings.Split(Helpers.Getenv("FILE_TYPES_ALLOWED"), "")
	mimeTypes = append(mimeTypes, exploded...)

	// Max file upload size is set to 10 mb
	var maxUploadSize int64
	if max, err := strconv.Atoi(Helpers.Getenv("FILE_MAX_UPLOAD_SIZE")); err != nil {
		maxUploadSize = 10 << 20
	} else {
		maxUploadSize = int64(max)
	}

	return &helpers.Helpers{
		Redner: a.Render,
		FileUploadConfig: helpers.FileUploadConfig{
			MaxSize:          maxUploadSize,
			AllowedMimeTypes: mimeTypes,
			TempDir:          "/storage/tmp",
			Destination:      "/storage/uploads",
		},
	}
}

// Configure the mailer for the application by initializing mailer struct. The mailer
// values are populated by the environemnt variables parsed from the .env file
// at the root of the application.
func (a *Adele) BoootstrapMailer() mailer.Mail {
	port, _ := strconv.Atoi(Helpers.Getenv("SMTP_PORT", "1025"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   a.RootPath + "/resources/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("MAILER_FROM_NAME"),
		FromAddress: os.Getenv("MAILER_FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}

	return m
}

// Configure the middleware for the application by initializing a middleware struct,
// populating its values using the application configuration.
func (a *Adele) BootstrapMiddleware() {
	myMiddleware := middleware.Middleware{
		FrameworkVersion: a.Version,
		AppName:          a.AppName,
		RootPath:         a.RootPath,
		Log:              a.Log,
		Session:          a.Session,
		MaintenanceMode:  a.MaintenanceMode,
	}

	a.middleware = myMiddleware
}

// Initializes a cron job scheduler for the Adele framework. Sets up task scheduling capabilities
// during application startup for framework-wide access.
func (a *Adele) BootstrapScheduler() {
	a.Scheduler = cron.New()
}

// Configure and create the session manager by initializing a session struct, populating
// its cookie fields by retrieving values from environment variables.
func (a *Adele) BootstrapSessionManager() (*scs.SessionManager, error) {

	session := session.Session{
		CookieDomain:   Helpers.Getenv("COOKIE_DOMAIN", "localhost"),
		CookieLifetime: Helpers.Getenv("COOKIE_LIFETIME", "1"),
		CookieName:     Helpers.Getenv("COOKIE_NAME", "adele"),
		CookiePersist:  Helpers.Getenv("COOKIE_PERSIST", "true"),
		CookieSecure:   Helpers.Getenv("COOKIE_SECURE", "false"),
	}

	switch strings.ToLower(os.Getenv("SESSION_TYPE")) {
	case "redis":
		//...

	case "mysql", "postgres", "mariadb", "postgresql":
		//...
	default:
		a.Log.Warn("sessions using in-memory session store")
	}

	manager := session.InitSession()
	return manager, nil
}

// Setup This code is setting up the Jet template engine for your Adele framework with
// different configurations based on whether the application is in debug/development mode
// or production mode—enables features that help during development but would hurt performance
// in production (like not caching templates and reloading them on every request).
func (a *Adele) BootstrapJetEngine() *jet.Set {
	loader := jet.NewOSFileSystemLoader(fmt.Sprintf("%s/%s", a.RootPath, a.ViewsTemplateDir))

	var views *jet.Set
	if a.Debug {
		views = jet.NewSet(loader, jet.InDevelopmentMode())
	} else {
		views = jet.NewSet(loader)
	}

	views.AddGlobal("APP_DEBUG", a.Debug)

	return views
}

// Setup up and configures an HTTP router using the adele mux package. This
// function returns an http.Handler which represents a chain of middleware
// and eventually, the handlers for specific routes.
func (a *Adele) BootstrapMux(rootPath string) (http.Handler, error) {

	// Load the applciation CORS (Cross-Origin Resource Sharing) configuration
	// from a YAML file and returning it as a mux.Cors object.
	configFile, err := os.ReadFile(fmt.Sprintf("%s/config/cors.yml", rootPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read cors config file: %v", err)
	}

	var corsConfig mux.Cors
	err = yaml.Unmarshal(configFile, &corsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	mux := mux.NewRouter()
	mux.Use(middleware.RequestID())
	mux.Use(middleware.RealIP())
	mux.Use(a.middleware.RateLimiter())

	corsOptions := crs.Options{
		AllowedOrigins:   corsConfig.AllowedOrigins,
		AllowedMethods:   corsConfig.AllowedMethods,
		AllowedHeaders:   corsConfig.AllowedHeaders,
		ExposedHeaders:   corsConfig.ExposedHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           corsConfig.MaxAge,
	}

	mux.Use(crs.Handler(corsOptions))

	if a.Debug {
		mux.Use(logger.HttpRequesLogger(logger.CreateLogger()))
		mux.Use(a.middleware.RecovererWithDebug)
	} else {
		mux.Use(middleware.Recoverer())
	}

	mux.Use(a.middleware.SessionLoad)
	mux.Use(a.middleware.CheckForMaintenanceMode)

	return mux, nil
}

// Setup and configuring a render engine- initializes a rendering system that handles
// template rendering for web responses (HTML pages, emails, etc.). The render system
// handles Rendering HTML templates for web pages, passing session data to templates,
// and, managing template inheritance and layouts.
func (a *Adele) BootstrapRender() *render.Render {
	r := render.Render{
		Directory: a.ViewsTemplateDir,
		Renderer:  a.config.renderer,
		RootPath:  a.RootPath,
		Port:      a.config.port,
		JetViews:  a.JetViews,
		Session:   a.Session,
	}

	return &r
}

// Cache initialization method that automatically detects and configures the appropriate
// caching system during application startup based on environment variables.
func (a *Adele) BootstrapCache(rootPath string) error {
	if cache.UsesRedis() {
		pool, err := redisdriver.CreateRedisPool(Helpers.Getenv("REDIS_MAX_IDEL", "50"), Helpers.Getenv("REDIS_MAX_ACTIVE_CONNECTIONS", "10000"), Helpers.Getenv("REDIS_TIMEOUT", "240"), Helpers.Getenv("REDIS_HOST", "localhost"), Helpers.Getenv("REDIS_PORT", "6380"))
		if err != nil {
			return err
		}

		rc := redisdriver.RedisCache{
			Conn:   pool,
			Prefix: Helpers.Getenv("REDIS_PREFIX", Helpers.Getenv("APP_NAME")),
		}

		a.Cache = &rc

	}

	if cache.UsesBadger() {

		bc := badgerdriver.BadgerCache{
			Conn: badgerdriver.CreateBadgerPool(a.RootPath + "/resources/badger"),
		}

		a.Cache = &bc

		a.Scheduler.AddFunc("@daily", func() {
			if err := badgerdriver.BadgerCacheClean(&bc); err != nil {
				a.Log.Errorf("Badger cache cleanup failed: %v", err)
			}
		})
	}

	return nil
}

// Ensure that a environment file at a specific path exists, creating it if it's missing, and returning
// any errors that may arise.
func (a *Adele) CreateEnvironmentFile(rootPath string) error {
	err := Helpers.CreateFileIfNotExist(fmt.Sprintf("%s/.env", rootPath))
	if err != nil {
		return err
	}
	return nil
}

// Create all nonexistent parent directories
func (a *Adele) CreateDirectories(rootPath string, directories []string) error {
	for _, path := range directories {
		err := Helpers.CreateDirIfNotExist(rootPath + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}
