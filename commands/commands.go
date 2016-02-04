package commands

import (
	"io/ioutil"
	"net/http"

	"github.com/RichardKnop/go-fixtures"
	"github.com/RichardKnop/recall/config"
	"github.com/RichardKnop/recall/health"
	"github.com/RichardKnop/recall/migrations"
	"github.com/RichardKnop/recall/oauth"
	"github.com/RichardKnop/recall/web"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/phyber/negroni-gzip/gzip"
)

// Migrate migrates the database
func Migrate(db *gorm.DB) error {
	// Bootsrrap migrations
	if err := migrations.Bootstrap(db); err != nil {
		return err
	}

	// Run migrations for the oauth service
	if err := oauth.MigrateAll(db); err != nil {
		return err
	}

	return nil
}

// LoadData loads fixtures
func LoadData(paths []string, cnf *config.Config, db *gorm.DB) error {
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if err := fixtures.Load(data, db.DB(), cnf.Database.Type); err != nil {
			return err
		}
	}

	return nil
}

// RunServer runs the app
func RunServer(cnf *config.Config, db *gorm.DB) {
	// Initialise the health service
	healthService := health.NewService(db)

	// Initialise the oauth service
	oauthService := oauth.NewService(cnf, db)

	// Initialise the web service
	webService := web.NewService(cnf, oauthService)

	// Start a classic negroni app
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(gzip.Gzip(gzip.DefaultCompression))
	app.Use(negroni.NewStatic(http.Dir("public")))

	// Create a router instance
	router := mux.NewRouter()

	// Add routes for the health service (healthcheck endpoint)
	health.RegisterRoutes(router, healthService)

	// Add routes for the oauth service (REST tokens endpoint)
	oauth.RegisterRoutes(router, oauthService)

	// Add routes for the web package (register, login authorize web pages)
	web.RegisterRoutes(router, webService)

	// Set the router
	app.UseHandler(router)

	// Run the server on port 8080
	app.Run(":8080")
}
