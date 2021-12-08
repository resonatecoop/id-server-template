package config

type ClientConfig struct {
	ConnectUrl  string `json:"connectUrl"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CSRFConfig struct {
	Key     string
	Origins string
}

type MailgunConfig struct {
	Sender string
	Key    string
	Domain string
}

// DatabaseConfig stores database connection options
type DatabaseConfig struct {
	PSN          string
	MaxIdleConns int
	MaxOpenConns int
}

// OauthConfig stores oauth service configuration options
type OauthConfig struct {
	AccessTokenLifetime  int
	RefreshTokenLifetime int
	AuthCodeLifetime     int
}

// SessionConfig stores session configuration for the web app
type SessionConfig struct {
	Secret string
	Domain string
	Path   string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge int
	Secure bool
	// When you tag a cookie with the HttpOnly flag, it tells the browser that
	// this particular cookie should only be accessed by the server.
	// Any attempt to access the cookie from client script is strictly forbidden.
	HTTPOnly bool
	SameSite bool
}

type Product struct {
	ID          string
	PriceID     string
	Name        string
	Description string
	Quantity    int64
}

type StripeConfig struct {
	Domain               string
	Token                string
	Secret               string
	WebHookSecret        string
	ListenerSubscription Product
	SupporterShares      Product
	ArtistMembership     Product
	LabelMembership      Product
	StreamCredit50       Product
	StreamCredit20       Product
	StreamCredit10       Product
	StreamCredit5        Product
}

// Config stores all configuration options
type Config struct {
	Hostname            string
	CSRF                CSRFConfig
	Mailgun             MailgunConfig
	Database            DatabaseConfig
	Oauth               OauthConfig
	Session             SessionConfig
	IsDevelopment       bool
	Clients             []ClientConfig
	Port                string
	ApplicationURL      string
	Origins             []string
	EmailTokenSecretKey string
	UserAPIHostname     string
	UserAPIPort         string
	StaticURL           string
	AppURL              string
	Stripe              StripeConfig
}
