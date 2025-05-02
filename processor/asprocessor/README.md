# ASN Processor

## Description

The ASN processor `asprocessor` enhances the attributes of a span, log, or metric by appending information about the autonomous system to which the IP address belongs. To add AS information, the IP address must be included in the attributes specified by the `attributes` configuration option (e.g., [`client.address`](https://github.com/open-telemetry/semantic-conventions/blob/v1.29.0/docs/general/attributes.md#client-attributes) and [`source.address`](https://github.com/open-telemetry/semantic-conventions/blob/v1.29.0/docs/general/attributes.md#source) by default). By default, only the resource attributes will be modified. Please refer to [config.go](./config.go) for the config spec.

### Autonomous System Metadata

The following [resource attributes](./internal/convention/attributes.go) will be added if the corresponding information is found:

```
  * as.number
  * as.organization.name
```

## Configuration

The following settings can be configured:

- `providers`: A map containing geographical location information providers. These providers are used to search for the geographical location attributes associated with an IP. Supported providers:
  - [maxmind](./internal/provider/maxmindprovider/README.md)
- `context` (default: `resource`): Allows specifying the underlying telemetry context the processor will work with. Available values:
  - `resource`: Resource attributes.
  - `record`: Attributes within a data point, log record or a span.
- `attributes` (default: `[client.address, source.address]`): An array of attribute names, which are used for the IP address lookup.

## Examples

```yaml
processors:
    as:
      providers:
        maxmind:
          database_path: /tmp/myasndb
      context: record
      attributes: [client.address, source.address, custom.address]
```
