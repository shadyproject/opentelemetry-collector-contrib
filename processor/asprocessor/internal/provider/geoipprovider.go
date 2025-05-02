// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package provider // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/provider"

import (
	"context"
	"errors"
	"net"

	"go.opentelemetry.io/collector/confmap/xconfmap"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/otel/attribute"
)

// ErrNoMetadataFound error should be returned when a provider could not find the corresponding ASN metadata
var ErrNoMetadataFound = errors.New("no ASN metadata found")

// Config is the configuration of a GeoIPProvider.
type Config interface {
	xconfmap.Validator
}

// AsProvider defines methods for obtaining the autonomous system information for a network
type AsProvider interface {
	// AutonomousSystem returns a set of attributes representing the Autonomous System information for a given IP address. It requires a context for managing request lifetime.
	AutonomousSystem(context.Context, net.IP) (attribute.Set, error)
}

// GeoIPProviderFactory can create GeoIPProvider instances.
type AsProviderFactory interface {
	// CreateDefaultConfig creates the default configuration for the GeoIPProvider.
	CreateDefaultConfig() Config

	// CreateAsProvider creates a provider based on this config. Processor's settings are provided as an argument to initialize the logger if needed.
	CreateAsProvider(ctx context.Context, settings processor.Settings, cfg Config) (AsProvider, error)
}
