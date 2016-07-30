package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

var (
	etcdEndpoint                          = "http://localhost:2379"
	etcdCertFile, etcdKeyFile, etcdCaFile string
	etcdConfigPath                        = "/config/example_api.json"
	configLoaded                          bool
)

// Cnf ...
// Let's start with some sensible defaults
var Cnf = &Config{
	Database: DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		User:         "example_api",
		Password:     "",
		DatabaseName: "example_api",
		MaxIdleConns: 5,
		MaxOpenConns: 5,
	},
	Oauth: OauthConfig{
		AccessTokenLifetime:  3600,    // 1 hour
		RefreshTokenLifetime: 1209600, // 14 days
		AuthCodeLifetime:     3600,    // 1 hour
	},
	Facebook: FacebookConfig{
		AppID:     "facebook_app_id",
		AppSecret: "facebook_app_secret",
	},
	Sendgrid: SendgridConfig{
		APIKey: "sendgrid_api_key",
	},
	Web: WebConfig{
		Scheme:    "http",
		Host:      "localhost:8080",
		AppScheme: "http",
		AppHost:   "localhost:8000",
	},
	AppSpecific: AppSpecificConfig{
		PasswordResetLifetime: 604800, // 7 days
		CompanyName:           "Your Company Name",
		CompanyEmail:          "contact@example.com",
	},
	IsDevelopment: true,
}

func init() {
	// Overwrite default values with environment variables if they are set
	if os.Getenv("ETCD_ENDPOINT") != "" {
		etcdEndpoint = os.Getenv("ETCD_ENDPOINT")
	}
	if os.Getenv("ETCD_CERT_FILE") != "" {
		etcdCertFile = os.Getenv("ETCD_CERT_FILE")
	}
	if os.Getenv("ETCD_KEY_FILE") != "" {
		etcdKeyFile = os.Getenv("ETCD_KEY_FILE")
	}
	if os.Getenv("ETCD_CA_FILE") != "" {
		etcdCaFile = os.Getenv("ETCD_CA_FILE")
	}
	if os.Getenv("ETCD_CONFIG_PATH") != "" {
		etcdConfigPath = os.Getenv("ETCD_CONFIG_PATH")
	}
}

// NewConfig loads configuration from etcd and returns *Config struct
// It also starts a goroutine in the background to keep config up-to-date
func NewConfig(mustLoadOnce bool, keepReloading bool) *Config {
	if configLoaded {
		return Cnf
	}

	// Init ETCD client
	etcdClient, err := newEtcdClient(etcdEndpoint, etcdCertFile, etcdKeyFile, etcdCaFile)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	// ETCD keys API
	kapi := client.NewKeysAPI(*etcdClient)

	// If the config must be loaded once successfully
	if mustLoadOnce {
		// Read from remote config the first time
		newCnf, err := LoadConfig(kapi)
		if err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}

		// Refresh the config
		RefreshConfig(newCnf)

		// Set configLoaded to true
		configLoaded = true
		logger.Info("Successfully loaded config for the first time")
	}

	if keepReloading {
		// Open a goroutine to watch remote changes forever
		go func() {
			for {
				// Delay after each request
				time.Sleep(time.Second * 10)

				// Attempt to reload the config
				newCnf, err := LoadConfig(kapi)
				if err != nil {
					logger.Error(err)
					continue
				}

				// Refresh the config
				RefreshConfig(newCnf)

				// Set configLoaded to true
				configLoaded = true
				logger.Info("Successfully reloaded config")
			}
		}()
	}

	return Cnf
}

// LoadConfig gets the JSON from ETCD and unmarshals it to the config object
func LoadConfig(kapi client.KeysAPI) (*Config, error) {
	// Read from remote config the first time
	resp, err := kapi.Get(context.Background(), etcdConfigPath, nil)
	if err != nil {
		return nil, err
	}

	// Unmarshal the config JSON into the cnf object
	newCnf := new(Config)
	if err := json.Unmarshal([]byte(resp.Node.Value), newCnf); err != nil {
		return nil, err
	}

	return newCnf, nil
}

// RefreshConfig sets config through the pointer so config actually gets refreshed
func RefreshConfig(newCnf *Config) {
	*Cnf = *newCnf
}

func newEtcdClient(theEndpoint, certFile, keyFile, caFile string) (*client.Client, error) {
	// Log the etcd endpoint for debugging purposes
	logger.Infof("ETCD Endpoint: %s", etcdEndpoint)

	// Start with the default HTTP transport
	var transport = client.DefaultTransport

	// Optionally, configure TLS transport
	if certFile != "" && keyFile != "" && caFile != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		// Load CA cert
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	// ETCD config
	etcdClientConfig := client.Config{
		Endpoints: []string{theEndpoint},
		Transport: transport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	// ETCD client
	etcdClient, err := client.New(etcdClientConfig)
	if err != nil {
		return nil, err
	}

	return &etcdClient, nil
}
