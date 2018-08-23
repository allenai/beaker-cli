package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClientAddress(t *testing.T) {
	cases := map[string]struct {
		address   string
		expected  string
		expectErr bool
	}{
		"Localhost": {
			address:  "localhost",
			expected: "https://localhost",
		},
		"IPv4": {
			address:  "127.0.0.1",
			expected: "https://127.0.0.1",
		},
		"IPv6": {
			address:  "[::1]",
			expected: "https://[::1]",
		},
		"SchemeAndHost": {
			address:  "tcp://example.com",
			expected: "tcp://example.com",
		},
		"Host": {
			address:  "example.com",
			expected: "https://example.com",
		},
		"HostAndPort": {
			address:  "example.com:80",
			expected: "https://example.com:80",
		},
		"TheWorks": {
			address:  "tcp://example.com:443",
			expected: "tcp://example.com:443",
		},
		"BadIPv6": {
			address:   "::1",
			expectErr: true,
		},
		"Subpath": {
			address:   "https://example.com:443/path",
			expectErr: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			client, err := NewClient(c.address, "")
			if c.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expected, client.Address())
			}
		})
	}
}
