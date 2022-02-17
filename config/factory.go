package config

import (
	"os"
	"time"

	"github.com/resonatecoop/id/log"
)

var (
	configLoaded   bool
	dialTimeout    = 5 * time.Second
	contextTimeout = 5 * time.Second
	reloadDelay    = time.Second * 10
)

// Cnf ...
// Let's start with some sensible defaults
var Cnf = &Config{
	Hostname: "id.resonate.coop",
	CSRF: CSRFConfig{
		Key:     "",
		Origins: "upload.resonate.is",
	},
	Mailgun: MailgunConfig{
		Sender: "members@resonate.is",
		Key:    "",
		Domain: "mailgun.resonate.is",
	},
	Database: DatabaseConfig{
		PSN:          "postgres://resonate_dev_user:password@127.0.0.1:5432/resonate_dev?sslmode=disable",
		MaxIdleConns: 5,
		MaxOpenConns: 5,
	},
	Oauth: OauthConfig{
		AccessTokenLifetime:  3600,    // 1 hour
		RefreshTokenLifetime: 1209600, // 14 days
		AuthCodeLifetime:     3600,    // 1 hour
	},
	Session: SessionConfig{
		Secret:   "test_secret",
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HTTPOnly: true,
	},
	Clients: []ClientConfig{
		{
			ConnectUrl:  "https://upload.resonate.is/api/user/connect/resonate",
			Name:        "Upload Tool",
			Description: "for creators",
		},
	},
	IsDevelopment:       true,
	Port:                ":8080",
	ApplicationURL:      "https://upload.resonate.is",
	Origins:             []string{"upload.resonate.is", "beta.stream.resonate.is"},
	EmailTokenSecretKey: "super secret key",
	UserAPIHostname:     "0.0.0.0",
	UserAPIPort:         ":11000",
	StaticURL:           "https://dash.resonate.coop",
	AppURL:              "https://stream.resonate.coop",
	Stripe: StripeConfig{
		WebHookSecret: "wh_",
		Domain:        "id.resonate.coop",
		Secret:        "sk_test_xxx",
		Token:         "pk_test_xxx",
		StreamCredit5: Product{
			ID:      "",
			PriceID: "price_xx",
		},
		StreamCredit10: Product{
			ID:      "",
			PriceID: "price_xx",
		},
		StreamCredit20: Product{
			ID:      "",
			PriceID: "price_xx",
		},
		StreamCredit50: Product{
			ID:      "",
			PriceID: "price_xx",
		},
		ListenerSubscription: Product{
			ID:       "",
			PriceID:  "price_xx",
			Quantity: int64(1),
		},
		SupporterShares: Product{
			ID:       "",
			PriceID:  "price_xx",
			Quantity: int64(0),
		},
		ArtistMembership: Product{
			ID:       "",
			PriceID:  "price_xx",
			Quantity: int64(1),
		},
		LabelMembership: Product{
			ID:       "",
			PriceID:  "price_xx",
			Quantity: int64(1),
		},
	},
}

// NewConfig loads configuration from etcd and returns *Config struct
// It also starts a goroutine in the background to keep config up-to-date
func NewConfig(mustLoadOnce bool, keepReloading bool, backendType string) *Config {
	if configLoaded {
		return Cnf
	}

	var backend Backend

	switch backendType {
	case "etcd":
		backend = new(etcdBackend)
	case "consul":
		backend = new(consulBackend)
	default:
		log.FATAL.Printf("%s is not a valid backend", backendType)
		os.Exit(1)
	}

	backend.InitConfigBackend()

	// If the config must be loaded once successfully
	if mustLoadOnce && !configLoaded {
		// Read from remote config the first time
		newCnf, err := backend.LoadConfig()

		if err != nil {
			log.FATAL.Print(err)
			os.Exit(1)
		}

		// Refresh the config
		backend.RefreshConfig(newCnf)

		// Set configLoaded to true
		configLoaded = true
		log.INFO.Print("Successfully loaded config for the first time")
	}

	if keepReloading {
		// Open a goroutine to watch remote changes forever
		go func() {
			for {
				// Delay after each request
				<-time.After(reloadDelay)

				// Attempt to reload the config
				newCnf, err := backend.LoadConfig()
				if err != nil {
					log.ERROR.Print(err)
					continue
				}

				// Refresh the config
				backend.RefreshConfig(newCnf)

				// Set configLoaded to true
				configLoaded = true
				log.INFO.Print("Successfully reloaded config")
			}
		}()
	}

	return Cnf
}
