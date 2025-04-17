// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package asprocessor

import (
	"context"
	"net"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processortest"
	"go.opentelemetry.io/otel/attribute"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"
	conventions "github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/convention"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/asprocessor/internal/provider"
)

type providerConfigMock struct {
	ValidateF func() error
}

type providerFactoryMock struct {
	CreateDefaultConfigF func() provider.Config
	CreateAsProviderF    func(context.Context, processor.Settings, provider.Config) (provider.AsProvider, error)
}

type providerMock struct {
	AutonomousSystemF func(context.Context, net.IP) (attribute.Set, error)
}

var (
	_ provider.AsProvider        = (*providerMock)(nil)
	_ provider.AsProvider        = (*providerMock)(nil)
	_ provider.AsProviderFactory = (*providerFactoryMock)(nil)
)

func (cm *providerConfigMock) Validate() error {
	return cm.ValidateF()
}

func (fm *providerFactoryMock) CreateDefaultConfig() provider.Config {
	return fm.CreateDefaultConfigF()
}

func (fm *providerFactoryMock) CreateAsProvider(ctx context.Context, settings processor.Settings, cfg provider.Config) (provider.AsProvider, error) {
	return fm.CreateAsProviderF(ctx, settings, cfg)
}

func (pm *providerMock) AutonomousSystem(ctx context.Context, ip net.IP) (attribute.Set, error) {
	return pm.AutonomousSystemF(ctx, ip)
}

var baseMockProvider = providerMock{
	AutonomousSystemF: func(context.Context, net.IP) (attribute.Set, error) {
		return attribute.Set{}, nil
	},
}

var baseMockFactory = providerFactoryMock{
	CreateDefaultConfigF: func() provider.Config {
		return &providerConfigMock{ValidateF: func() error { return nil }}
	},
	CreateAsProviderF: func(context.Context, processor.Settings, provider.Config) (provider.AsProvider, error) {
		return &baseMockProvider, nil
	},
}

var baseProviderMock = providerMock{
	AutonomousSystemF: func(context.Context, net.IP) (attribute.Set, error) {
		return attribute.Set{}, nil
	},
}

var testCases = []struct {
	name       string
	goldenDir  string
	context    ContextID
	attributes []attribute.Key
}{
	{
		name:      "default source.address attribute, not found",
		goldenDir: "resource_no_source_address",
		context:   resource,
	},
	{
		name:      "default source.address attribute",
		goldenDir: "resource_source_address",
		context:   resource,
	},
	{
		name:      "default source.address attribute no geo metadata found by providers",
		goldenDir: "resource_source_address_geo_not_found",
		context:   resource,
	},
	{
		name:      "default source.ip attribute with an unspecified IP address should be skipped",
		goldenDir: "resource_unspecified_address",
		context:   resource,
	},
	{
		name:      "do not add resource attributes with an invalid ip",
		goldenDir: "resource_invalid_address",
		context:   resource,
	},
	{
		name:      "source address located in the record attributes",
		goldenDir: "record_source_address",
		context:   record,
	},
	{
		name:      "client address located in the record attributes",
		goldenDir: "record_client_address",
		context:   record,
	},
	{
		name:       "custom address located in the record attributes",
		goldenDir:  "record_custom_address",
		context:    record,
		attributes: []attribute.Key{"source.address", "client.address", "custom.address"},
	},
}

func compareAllSignals(cfg component.Config, goldenDir string) func(t *testing.T) {
	return func(t *testing.T) {
		dir := filepath.Join("testdata", goldenDir)
		factory := NewFactory()

		// compare metrics
		nextMetrics := new(consumertest.MetricsSink)
		metricsProcessor, err := factory.CreateMetrics(context.Background(), processortest.NewNopSettings(metadata.Type), cfg, nextMetrics)
		require.NoError(t, err)

		inputMetrics, err := golden.ReadMetrics(filepath.Join(dir, "input-metrics.yaml"))
		require.NoError(t, err)

		expectedMetrics, err := golden.ReadMetrics(filepath.Join(dir, "output-metrics.yaml"))
		require.NoError(t, err)

		err = metricsProcessor.ConsumeMetrics(context.Background(), inputMetrics)
		require.NoError(t, err)

		actualMetrics := nextMetrics.AllMetrics()
		require.Len(t, actualMetrics, 1)
		// golden.WriteMetrics(t, filepath.Join(dir, "output-metrics.yaml"), actualMetrics[0])
		require.NoError(t, pmetrictest.CompareMetrics(expectedMetrics, actualMetrics[0]))

		// compare traces
		nextTraces := new(consumertest.TracesSink)
		tracesProcessor, err := factory.CreateTraces(context.Background(), processortest.NewNopSettings(metadata.Type), cfg, nextTraces)
		require.NoError(t, err)

		inputTraces, err := golden.ReadTraces(filepath.Join(dir, "input-traces.yaml"))
		require.NoError(t, err)

		expectedTraces, err := golden.ReadTraces(filepath.Join(dir, "output-traces.yaml"))
		require.NoError(t, err)

		err = tracesProcessor.ConsumeTraces(context.Background(), inputTraces)
		require.NoError(t, err)

		actualTraces := nextTraces.AllTraces()
		require.Len(t, actualTraces, 1)
		// golden.WriteTraces(t, filepath.Join(dir, "output-traces.yaml"), actualTraces[0])
		require.NoError(t, ptracetest.CompareTraces(expectedTraces, actualTraces[0]))

		// compare logs
		nextLogs := new(consumertest.LogsSink)
		logsProcessor, err := factory.CreateLogs(context.Background(), processortest.NewNopSettings(metadata.Type), cfg, nextLogs)
		require.NoError(t, err)

		inputLogs, err := golden.ReadLogs(filepath.Join(dir, "input-logs.yaml"))
		require.NoError(t, err)

		err = logsProcessor.ConsumeLogs(context.Background(), inputLogs)
		require.NoError(t, err)

		expectedLogs, err := golden.ReadLogs(filepath.Join(dir, "output-logs.yaml"))
		require.NoError(t, err)

		actualLogs := nextLogs.AllLogs()
		require.Len(t, actualLogs, 1)
		// golden.WriteLogs(t, filepath.Join(dir, "output-logs.yaml"), actualLogs[0])
		require.NoError(t, plogtest.CompareLogs(expectedLogs, actualLogs[0]))
	}
}

func TestProcessor(t *testing.T) {
	t.Parallel()

	baseMockFactory.CreateAsProviderF = func(context.Context, processor.Settings, provider.Config) (provider.AsProvider, error) {
		return &baseProviderMock, nil
	}

	baseProviderMock.AutonomousSystemF = func(_ context.Context, sourceIP net.IP) (attribute.Set, error) {
		if sourceIP.Equal(net.IPv4(1, 2, 3, 4)) {
			return attribute.NewSet([]attribute.KeyValue{
				attribute.Int(conventions.AttributeAsNumber, 13337),
				attribute.String(conventions.AttributesAsOrganizationName, "CLOUDFLARENET"),
			}...), nil
		}
		return attribute.Set{}, provider.ErrNoMetadataFound
	}
	const providerKey string = "mock"
	providerFactories[providerKey] = &baseMockFactory

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var attributes []attribute.Key = defaultAttributes
			if tt.attributes != nil {
				attributes = tt.attributes
			}
			cfg := &Config{Context: tt.context, Providers: map[string]provider.Config{providerKey: &providerConfigMock{}}, Attributes: attributes}
			compareAllSignals(cfg, tt.goldenDir)(t)
		})
	}
}
