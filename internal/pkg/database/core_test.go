package database

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildConnectionString(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name                 string
		connectionStringBase string
		params               map[string]string
		expected             string
		description          string
	}{
		{
			name:                 "TCP connection with standard parameters",
			connectionStringBase: "postgres://localhost:5432/mydb",
			params: map[string]string{
				"dbname":  "mydb",
				"sslmode": "disable",
				"user":    "postgres",
			},
			expected:    "postgres://localhost:5432/mydb?dbname=mydb&sslmode=disable&user=postgres",
			description: "Standard TCP connection should use base connection string with encoded parameters",
		},
		{
			name:                 "Unix socket connection with socket path",
			connectionStringBase: "postgres:///mydb",
			params: map[string]string{
				"dbname":  "mydb",
				"host":    "/var/run/postgresql",
				"sslmode": "disable",
			},
			expected:    "postgres:///mydb?dbname=mydb&host=%2Fvar%2Frun%2Fpostgresql&sslmode=disable",
			description: "Unix socket connection should be detected by host parameter starting with '/'",
		},
		{
			name:                 "Unix socket connection with different database name",
			connectionStringBase: "postgres:///postgres",
			params: map[string]string{
				"dbname": "testdb",
				"host":   "/tmp/postgresql",
			},
			expected:    "postgres:///testdb?dbname=testdb&host=%2Ftmp%2Fpostgresql",
			description: "Unix socket should use dbname parameter for database in URL",
		},
		{
			name:                 "Unix socket connection without dbname defaults to postgres",
			connectionStringBase: "postgres:///",
			params: map[string]string{
				"host":    "/var/run/postgresql",
				"sslmode": "disable",
			},
			expected:    "postgres:///postgres?host=%2Fvar%2Frun%2Fpostgresql&sslmode=disable",
			description: "Unix socket without dbname should default to 'postgres' database",
		},
		{
			name:                 "TCP connection with host parameter not starting with slash",
			connectionStringBase: "postgres://localhost:5432/mydb",
			params: map[string]string{
				"dbname": "mydb",
				"host":   "localhost",
			},
			expected:    "postgres://localhost:5432/mydb?dbname=mydb&host=localhost",
			description: "Host parameter not starting with '/' should be treated as TCP connection",
		},
		{
			name:                 "TCP connection without host parameter",
			connectionStringBase: "postgres://dbhost:5432/mydb",
			params: map[string]string{
				"dbname":  "mydb",
				"sslmode": "require",
			},
			expected:    "postgres://dbhost:5432/mydb?dbname=mydb&sslmode=require",
			description: "Connection without host parameter should use standard TCP format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// Convert map to url.Values
			params := url.Values{}
			for k, v := range tt.params {
				params.Add(k, v)
			}

			result := buildConnectionString(tt.connectionStringBase, params)

			// Parse both URLs to compare them properly (order of query params may vary)
			expectedURL, err := url.Parse(tt.expected)
			assert.NoError(t, err, "Expected URL should be valid")

			resultURL, err := url.Parse(result)
			assert.NoError(t, err, "Result URL should be valid")

			// Compare scheme, host, path
			assert.Equal(t, expectedURL.Scheme, resultURL.Scheme, "Schemes should match")
			assert.Equal(t, expectedURL.Host, resultURL.Host, "Hosts should match")
			assert.Equal(t, expectedURL.Path, resultURL.Path, "Paths should match")

			// Compare query parameters (order independent)
			expectedQuery := expectedURL.Query()
			resultQuery := resultURL.Query()
			assert.Equal(t, expectedQuery, resultQuery, "Query parameters should match")
		})
	}
}
