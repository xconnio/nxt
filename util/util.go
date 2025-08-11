package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/xconnio/xconn-go"
)

func StartServerFromConfigFile(configFile string) ([]io.Closer, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	var decoder = yaml.NewDecoder(bytes.NewBuffer(data))
	decoder.KnownFields(true)

	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	router := xconn.NewRouter()
	defer router.Close()

	for _, realm := range config.Realms {
		router.AddRealm(realm.Name)
	}

	authenticator := NewAuthenticator(config.Authenticators)

	closers := make([]io.Closer, 0)
	for _, transport := range config.Transports {
		var throttle *xconn.Throttle
		if transport.RateLimit.Rate > 0 && transport.RateLimit.Interval > 0 {
			strategy := xconn.Burst
			if transport.RateLimit.Strategy == LeakyBucketStrategy {
				strategy = xconn.LeakyBucket
			}
			throttle = xconn.NewThrottle(transport.RateLimit.Rate,
				time.Duration(transport.RateLimit.Interval)*time.Second, strategy)
		}
		server := xconn.NewServer(router, authenticator, &xconn.ServerConfig{Throttle: throttle})

		var closer io.Closer
		switch transport.Type {
		case WebSocketTransport:
			closer, err = server.ListenAndServeWebSocket(transport.Listener, transport.Address)
			if err != nil {
				return nil, err
			}
			if transport.Listener == xconn.NetworkUnix {
				fmt.Printf("listening websocket on unix+ws://%s\n", transport.Address)
			} else {
				fmt.Printf("listening websocket on ws://%s\n", transport.Address)
			}
		case UniversalTransport:
			closer, err = server.ListenAndServeUniversal(transport.Listener, transport.Address)
			if err != nil {
				return nil, err
			}
			if transport.Listener == xconn.NetworkUnix {
				fmt.Printf("listening rawsocket on unix://%s\n", transport.Address)
				fmt.Printf("listening websocket on unix+ws://%s\n", transport.Address)
			} else {
				fmt.Printf("listening rawsocket on rs://%s\n", transport.Address)
				fmt.Printf("listening websocket on ws://%s\n", transport.Address)
			}
		case RawSocketTransport:
			closer, err = server.ListenAndServeRawSocket(transport.Listener, transport.Address)
			if err != nil {
				return nil, err
			}
			if transport.Listener == xconn.NetworkUnix {
				fmt.Printf("listening rawsocket on unix://%s\n", transport.Address)
			} else {
				fmt.Printf("listening rawsocket on rs://%s\n", transport.Address)
			}
		}

		closers = append(closers, closer)
	}

	return closers, nil
}
