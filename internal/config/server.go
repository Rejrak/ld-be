package config

import (
	"be/internal/utils"
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"

	"goa.design/clue/log"
)

// ServerConfig holds server-related configurations such as domain, host, port, and security settings.
type ServerConfig struct {
	Domain   string // The server domain, e.g., "development" or "production"
	Host     string // Host IP address or domain name for the server
	HTTPPort string // Port for the HTTP server
	Port     string // Optional custom port if needed for other services
	Secure   bool   // Flag to use secure HTTPS connections
	Debug    bool   // Debug mode flag to enable detailed logging
}

// LoadServerConfig parses command-line flags and loads environment variables as needed.
// This function sets up the server configuration based on the environment.
func LoadServerConfig() *ServerConfig {

	var (
		// Flag for selecting the host environment (development or production)
		domainF = flag.String("host", "development", "Server host (valid values: development, production)")
		// Flag for specifying the host domain or IP address
		hostF = flag.String("domain", "0.0.0.0", "Host domain name")
		// Flag for setting the HTTP port
		httpPortF = flag.String("http-port", "9090", "HTTP port")
		// Flag to enable secure connections (HTTPS)
		secureF = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		// Debug flag to log request and response bodies
		dbgF = flag.Bool("debug", true, "Log request and response bodies")
		// Flag to specify environment for loading the appropriate .env file
		envF = flag.String("env", "develop", "load .env when outside a docker container")
	)

	flag.Parse() // Parse command-line flags

	// Load environment variables based on the selected environment
	switch *envF {
	case "vpn":
		utils.Env.LoadEnv(".env") // Load .env file for VPN environment
	case "local":
		utils.Env.LoadEnv(".env_local") // Load .env_local file for local environment
	default:
		break // Panic if an unsupported environment is specified
	}

	// Return a ServerConfig instance populated with the parsed flag values
	return &ServerConfig{
		Host:     *hostF,
		Domain:   *domainF,
		HTTPPort: *httpPortF,
		Secure:   *secureF,
		Debug:    *dbgF,
	}
}

// BuildServerURL constructs a URL based on the ServerConfig settings.
// This method dynamically builds the URL based on host, port, and security settings.
func (sc *ServerConfig) BuildServerURL(config *ServerConfig, ctx context.Context) *url.URL {

	// Initialize URL with the HTTP scheme, host, and port from configuration
	addr := fmt.Sprintf("http://%s:%s", config.Host, config.HTTPPort)
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatalf(ctx, err, "invalid URL %#v\n", addr) // Log a fatal error if the URL is invalid
	}

	// If the Secure flag is set, switch the scheme to HTTPS
	if config.Secure {
		u.Scheme = "https"
	}

	// Set the host and port if the host is specified in configuration
	if config.Host != "" {
		u.Host = fmt.Sprintf("%s:%s", config.Host, config.HTTPPort)
	}

	// If HTTPPort is set, ensure the port is correctly formatted in the URL
	if config.HTTPPort != "" {
		h, _, err := net.SplitHostPort(u.Host) // Split the host and port components
		if err != nil {
			log.Fatalf(ctx, err, "invalid URL %#v\n", u.Host) // Log error if splitting fails
		}
		u.Host = net.JoinHostPort(h, config.HTTPPort) // Rejoin with the specified port
	} else if u.Port() == "" {
		u.Host = net.JoinHostPort(u.Host, "80") // Default to port 80 if no port is specified
	}

	return u // Return the constructed URL
}
