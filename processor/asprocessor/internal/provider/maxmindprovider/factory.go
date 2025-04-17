// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package maxmind // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/geoipprocessor/internal/provider/maxmindprovider"

import (
	"context"

	"go.opentelemetry.io/collector/processor"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/provider"
)

const (
	// TypeStr the value of "type" key in configuration.
	TypeStr = "maxmind"
)

// Factory is the Factory for the MaxMind GeoIP provider.
type Factory struct{}

var _ provider.AsProviderFactory = (*Factory)(nil)

// CreateDefaultConfig creates the default configuration for the Provider.
func (f *Factory) CreateDefaultConfig() provider.Config {
	return &Config{}
}

// CreateAsProvider creates a provider based on this config.
func (f *Factory) CreateAsProvider(_ context.Context, _ processor.Settings, cfg provider.Config) (provider.AsProvider, error) {
	maxMindConfig := cfg.(*Config)
	return newMaxMindProvider(maxMindConfig)
}
