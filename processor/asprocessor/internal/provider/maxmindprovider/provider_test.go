// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package maxmind

import (
	"context"
	"net"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	conventions "github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/convention"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/provider/maxmindprovider/testdata"
)

func TestInvalidNewProvider(t *testing.T) {
	_, err := newMaxMindProvider(&Config{})
	expectedErrMsgSuffix := "no such file or directory"
	if runtime.GOOS == "windows" {
		expectedErrMsgSuffix = "The system cannot find the file specified."
	}
	require.ErrorContains(t, err, "could not open asn database: open : "+expectedErrMsgSuffix)

	_, err = newMaxMindProvider(&Config{DatabasePath: "no valid path"})
	require.ErrorContains(t, err, "could not open asn database: open no valid path: "+expectedErrMsgSuffix)
}

// TestProviderLocation asserts that the MaxMind provider adds the geo location data given an IP.
func TestProviderLocation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38961")
	}
	tmpDBfiles := testdata.GenerateLocalDB(t, "./testdata")
	defer os.RemoveAll(tmpDBfiles)

	t.Parallel()

	tests := []struct {
		name               string
		testDatabase       string
		sourceIP           net.IP
		expectedAttributes attribute.Set
		expectedErrMsg     string
	}{
		{
			name:           "nil IP address",
			testDatabase:   "GeoIP2-ASN-Test.mmdb",
			expectedErrMsg: "IP passed to Lookup cannot be nil",
		},
		{
			name:           "unsupported database type",
			sourceIP:       net.IPv4(0, 0, 0, 0),
			testDatabase:   "GeoIP2-Domain-Test.mmdb",
			expectedErrMsg: "unsupported geo IP database type type: GeoIP2-Domain",
		},
		{
			// TODO: change then when we add test data
			name:           "no IP metadata in database",
			sourceIP:       net.IPv4(0, 0, 0, 0),
			testDatabase:   "GeoIP2-ISP-Test.mmdb",
			expectedErrMsg: "no geo IP metadata found",
		},
		{
			name:         "all attributes should be present for IPv4 using GeoLite2-City database",
			sourceIP:     net.IPv4(1, 2, 3, 4),
			testDatabase: "GeoLite2-ASN-Test.mmdb",
			expectedAttributes: attribute.NewSet([]attribute.KeyValue{
				attribute.Int(conventions.AttributeAsNumber, 13337),
				attribute.String(conventions.AttributesAsOrganizationName, "CLOUDFLARENET"),
			}...),
		},
		{
			name:         "subset attributes for IPv6 IP using GeoIP2-City database",
			sourceIP:     net.ParseIP("2001:220::"),
			testDatabase: "GeoIP2-ISP-Test.mmdb",
			expectedAttributes: attribute.NewSet([]attribute.KeyValue{
				attribute.Int(conventions.AttributeAsNumber, 13337),
				attribute.String(conventions.AttributesAsOrganizationName, "CLOUDFLARENET"),
			}...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare provider
			provider, err := newMaxMindProvider(&Config{DatabasePath: tmpDBfiles + "/" + tt.testDatabase})
			assert.NoError(t, err)

			// assert metrics
			actualAttributes, err := provider.AutonomousSystem(context.Background(), tt.sourceIP)
			if tt.expectedErrMsg != "" {
				assert.EqualError(t, err, tt.expectedErrMsg)
				return
			}

			assert.True(t, tt.expectedAttributes.Equals(&actualAttributes))
		})
	}
}
