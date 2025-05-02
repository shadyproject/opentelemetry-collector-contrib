# MaxMind GeoIP Provider

> Use of MaxMind and other geolocation databases are subject to applicable licenses and terms governing the databases. Consult the database provider for the latest applicable terms.

This package provides a MaxMind GeoIP provider for use with the OpenTelemetry GeoIP processor. It leverages the [geoip2-golang package](https://github.com/oschwald/geoip2-golang) to query autonomous system information associated with IP addresses from MaxMind databases. See recommended clients: https://dev.maxmind.com/geoip/docs/databases#api-clients

# Features

- Supports GeoIP2-ISP and GeoLite2-ASN database types.
- Retrieves and returns Autonomus System metadata for a given IP address. The generated attributes follow the internal [ASN conventions](../../convention/attributes.go).

## Configuration

The following configuration must be provided:

- `database_path`: local file path to a GeoIP2-ISP or GeoLite2-ASN database.
