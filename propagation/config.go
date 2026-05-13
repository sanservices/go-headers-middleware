package propagation

import (
	"os"
	"strings"
)

const DefaultEnvVar = "FORWARD_HEADERS"

type Config struct {
	Headers []string
}

// LoadConfig reads FORWARD_HEADERS and parses a comma-separated list.
func LoadConfig() Config {
	return LoadConfigFromEnv(DefaultEnvVar)
}

func LoadConfigFromEnv(envVar string) Config {
	value := os.Getenv(envVar)
	if value == "" {
		return Config{}
	}

	seen := make(map[string]struct{})
	headers := make([]string, 0)

	for _, h := range strings.Split(value, ",") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}

		// HTTP headers are case-insensitive.
		key := strings.ToLower(h)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		headers = append(headers, h)
	}

	return Config{Headers: headers}
}

func (c Config) Enabled() bool {
	return len(c.Headers) > 0
}
