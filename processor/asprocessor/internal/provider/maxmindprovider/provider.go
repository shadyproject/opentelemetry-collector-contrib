// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package maxmind // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/provider/maxmindprovider"

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
	"go.opentelemetry.io/otel/attribute"

	conventions "github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/convention"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/provider"
)

var (
	// defaultLanguageCode specifies English as the default Geolocation language code, see https://dev.maxmind.com/geoip/docs/web-services/responses#languages
	defaultLanguageCode = "en"
	geoIP2IspDBType     = "GeoIP2-ISP"
	geoLite2AsnDBType   = "GeoLite2-ASN"

	errUnsupportedDB = errors.New("unsupported geo IP database type")
)

type maxMindProvider struct {
	geoReader *geoip2.Reader
	// language code to be used in name retrieval, e.g. "en" or "pt-BR"
	langCode string
}

var _ provider.AsProvider = (*maxMindProvider)(nil)

func newMaxMindProvider(cfg *Config) (*maxMindProvider, error) {
	geoReader, err := geoip2.Open(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("could not open geoip database: %w", err)
	}

	return &maxMindProvider{geoReader: geoReader, langCode: defaultLanguageCode}, nil
}

// Location implements provider.GeoIPProvider for MaxMind. If a non City database type is used or no metadata is found in the database, an error will be returned.
func (g *maxMindProvider) AutonomousSystem(_ context.Context, ipAddress net.IP) (attribute.Set, error) {
	switch g.geoReader.Metadata().DatabaseType {
	case geoIP2IspDBType, geoLite2AsnDBType:
		attrs, err := g.asAttributes(ipAddress)
		if err != nil {
			return attribute.Set{}, err
		} else if len(*attrs) == 0 {
			return attribute.Set{}, provider.ErrNoMetadataFound
		}
		return attribute.NewSet(*attrs...), nil
	default:
		return attribute.Set{}, fmt.Errorf("%w type: %s", errUnsupportedDB, g.geoReader.Metadata().DatabaseType)
	}
}

// cityAttributes returns a list of key-values containing geographical metadata associated to the provided IP. The key names are populated using the internal geo IP conventions package. If an invalid or nil IP is provided, an error is returned.
func (g *maxMindProvider) asAttributes(ipAddress net.IP) (*[]attribute.KeyValue, error) {
	attributes := make([]attribute.KeyValue, 0, 11)

	asn, err := g.geoReader.ASN(ipAddress)
	if err != nil {
		return nil, err
	}

	// The exact set of top-level keys varies based on the particular GeoIP2 web service you are using. If a key maps to an undefined or empty value, it is not included in the JSON object. The following anonymous function appends the given key-value only if the value is not empty.
	appendIfNotEmpty := func(keyName, value string) {
		if value != "" {
			attributes = append(attributes, attribute.String(keyName, value))
		}
	}

	// AS number
	attributes = append(attributes, attribute.Int(conventions.AttributeAsNumber, int(asn.AutonomousSystemNumber)))
	appendIfNotEmpty(conventions.AttributesAsOrganizationName, asn.AutonomousSystemOrganization)

	// TODO: maxmind db asn has more geo info, possibly add it in here

	return &attributes, err
}
