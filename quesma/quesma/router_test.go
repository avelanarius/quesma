package quesma

import (
	"github.com/stretchr/testify/assert"
	"mitmproxy/quesma/quesma/config"
	"testing"
)

func Test_matchedAgainstConfig(t *testing.T) {
	tests := []struct {
		name   string
		index  string
		config config.QuesmaConfiguration
		want   bool
	}{
		{
			name:   "index enabled",
			index:  "index",
			config: indexConfig("index", true),
			want:   true,
		},
		{
			name:   "index enabled via pattern",
			index:  "index",
			config: indexConfig("ind*", true),
			want:   true,
		},
		{
			name:   "index enabled via complex pattern",
			index:  "index",
			config: indexConfig("i*d*x", true),
			want:   true,
		},
		{
			name:   "index disabled via complex pattern",
			index:  "index",
			config: indexConfig("i*d*x", false),
			want:   false,
		},
		{
			name:   "index disabled",
			index:  "index",
			config: indexConfig("index", false),
			want:   false,
		},
		{
			name:   "index not configured",
			index:  "index",
			config: indexConfig("logs-*", false),
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, matchedAgainstConfig(tt.config)(map[string]string{"index": tt.index}), "matchedAgainstConfig(%v), index: %s", tt.config, tt.index)
		})
	}
}

func Test_matchedAgainstPattern(t *testing.T) {
	tests := []struct {
		name          string
		index         string
		tables        []string
		configuration config.QuesmaConfiguration
		want          bool
	}{
		{
			name:          "index enabled and table present",
			index:         "index",
			tables:        []string{"index"},
			configuration: indexConfig("index", true),
			want:          true,
		},
		{
			name:          "index disabled and table present",
			index:         "index",
			tables:        []string{"index"},
			configuration: indexConfig("index", false),
			want:          false,
		},
		{
			name:          "index enabled and table not present",
			index:         "index",
			tables:        []string{},
			configuration: indexConfig("index", true),
			want:          false,
		},
		{
			name:          "index disabled and table not present",
			index:         "index",
			tables:        []string{},
			configuration: indexConfig("index", false),
			want:          false,
		},
		{
			name:          "index enabled, wide pattern, table present",
			index:         "logs-*-*",
			tables:        []string{"logs-generic-default"},
			configuration: indexConfig("logs-generic-*", true),
			want:          true,
		},
		{
			name:          "index enabled, wide pattern, table not present",
			index:         "logs-*-*",
			tables:        []string{},
			configuration: indexConfig("logs-generic-*", true),
			want:          false,
		},
		{
			name:          "index disabled, wide pattern, table present",
			index:         "logs-*-*",
			tables:        []string{"logs-generic-default"},
			configuration: indexConfig("logs-generic-*", false),
			want:          false,
		},
		{
			name:          "index enabled, same pattern, table present",
			index:         "logs-*-*",
			tables:        []string{"logs-generic-default"},
			configuration: indexConfig("logs-*-*", true),
			want:          true,
		},
		{
			name:          "index enabled, same pattern, table not present",
			index:         "logs-*-*",
			tables:        []string{},
			configuration: indexConfig("logs-*-*", true),
			want:          false,
		},
		{
			name:          "index disabled, same pattern, table present",
			index:         "logs-*-*",
			tables:        []string{"logs-generic-default"},
			configuration: indexConfig("logs-*-*", false),
			want:          false,
		},
		{
			name:          "index enabled, narrow pattern, table present",
			index:         "logs-generic-*",
			tables:        []string{"logs-generic-default"},
			configuration: indexConfig("logs-*", true),
			want:          true,
		},
		{
			name:          "index enabled, narrow pattern, table not present",
			index:         "logs-generic-*",
			tables:        []string{},
			configuration: indexConfig("logs-*", true),
			want:          false,
		},
		{
			name:          "index disabled, narrow pattern, table present",
			index:         "logs-generic-*",
			tables:        []string{"logs-generic-default"},
			configuration: indexConfig("logs-*", false),
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, matchedAgainstPattern(tt.configuration, func() []string {
				return tt.tables
			})(map[string]string{"index": tt.index}), "matchedAgainstPattern(%v)", tt.configuration)
		})
	}
}

func indexConfig(pattern string, enabled bool) config.QuesmaConfiguration {
	return config.QuesmaConfiguration{IndexConfig: []config.IndexConfiguration{{NamePattern: pattern, Enabled: enabled}}}
}