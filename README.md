# AppLovin Max MCP Server

A Model Context Protocol (MCP) server that provides LLM clients like Claude with access to AppLovin Max advertising platform analytics and reporting capabilities.

---

**⚠️ IMPORTANT: PROOF OF CONCEPT ONLY**

This is a proof-of-concept throwaway project created for demonstration and experimentation purposes. It is **NOT production-ready** and should **NOT be used in production environments** without significant additional work.

**Known Limitations:**
- No comprehensive error handling or recovery mechanisms
- No rate limiting or request throttling
- No input validation beyond basic parameter checks
- No logging or monitoring capabilities
- No unit tests or integration tests
- No security audits performed
- No performance optimization or scalability considerations
- API keys transmitted in query parameters (consider more secure methods for production)

**Use at your own risk.** This code is provided as-is with no warranties. See the LICENSE file for details.

---

## Overview

This Go-based MCP server exposes AppLovin Max API functionality through standardized tools, enabling AI assistants to query mobile ad revenue data, analyze user cohorts, and generate insights from your AppLovin monetization metrics.

## Features

- **Revenue Reporting**: Access aggregated mediation statistics including impressions, revenue, eCPM, and fill rates
- **Cohort Analysis**: Track user lifecycle metrics segmented by installation date
- **Flexible Filtering**: Filter by application, country, platform, ad type, and more
- **Multiple Data Views**: Choose between revenue, impression, or session-based cohort analysis
- **Time-Series Tracking**: Monitor how metrics evolve over time (up to 45 days post-install)

## Prerequisites

- Go 1.25.4 or later
- AppLovin Max account with API access
- AppLovin API key

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd Applovin-Max-MCP
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
go build -o bootstrap cmd/main.go
```

## Configuration

Set your AppLovin API key as an environment variable:

```bash
export APPLOVIN_API_KEY="your_api_key_here"
```

## Usage

### Starting the Server

The MCP server communicates via stdio (standard input/output):

```bash
./bootstrap
```

### Integrating with Claude Desktop

Add the following configuration to your Claude Desktop config file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "applovin-max": {
      "command": "/path/to/Applovin-Max-MCP/bootstrap",
      "env": {
        "APPLOVIN_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

## Available Tools

### 1. Revenue Report (`revenue_report`)

Retrieves aggregated mediation statistics from AppLovin Max.

#### Required Parameters

- `start` (string): Start date in YYYY-MM-DD format
- `end` (string): End date in YYYY-MM-DD format
- `format` (string): Output format - `json` or `csv`

#### Optional Parameters

**Columns** (array, default: `["day", "application"]`):
- **Dimensions**: `day`, `hour`, `application`, `package_name`, `ad_format`, `country`, `platform`, `network`, `network_placement`, `max_placement`, `max_ad_unit_id`, `custom_network_name`, `ad_unit_waterfall_name`, `device_type`, `store_id`, `has_idfa`, `max_ad_unit_test`
- **Metrics**: `impressions`, `responses`, `requests`, `attempts`, `estimated_revenue`, `ecpm`, `fill_rate`

**Filters**:
- `filter_application`: Filter by application name
- `filter_package_name`: Filter by package name
- `filter_ad_type`: Filter by ad type (e.g., banner, inter, rewarded)
- `filter_country`: Filter by two-letter ISO country code
- `filter_platform`: Filter by platform (android, ios)
- `filter_network`: Filter by network name
- `filter_zone`: Filter by zone

**Sorting**:
- `sort_day`: Sort by day (ASC/DESC)
- `sort_hour`: Sort by hour (ASC/DESC)
- `sort_estimated_revenue`: Sort by estimated revenue (ASC/DESC)

**Pagination**:
- `limit`: Limit number of results
- `offset`: Skip first N results

**Other**:
- `not_zero` (boolean): Exclude rows where all metrics are zero

#### Example Usage

```
Get revenue data for the last 7 days by country
```

Claude will use:
```json
{
  "start": "2025-11-25",
  "end": "2025-12-02",
  "format": "json",
  "columns": ["day", "country", "estimated_revenue", "impressions", "ecpm"]
}
```

### 2. Cohort Request (`cohort_request`)

Analyzes user cohort performance segmented by installation date.

#### Required Parameters

- `start` (string): Start date (installation date) in YYYY-MM-DD format
- `end` (string): End date in YYYY-MM-DD format (max 45-day range)
- `format` (string): Output format - `json` or `csv`

#### Optional Parameters

**Cohort Type** (string, default: `revenue`):
- `revenue`: Ad revenue and IAP metrics
- `impression`: Ad impression data
- `session`: Retention and session metrics

**Columns** (array, default: `["day", "installs"]`):
- **Dimensions**: `day`, `installs`, `country`, `platform`, `package_name`, `application`
- **Revenue Metrics**: `ads_rpi`, `iap_rpi`, `pub_revenue`, `inter_rpi`, `banner_rpi`, `reward_rpi`
- **Impression Metrics**: `imp`, `imp_per_user`, `banner_imp`, `inter_imp`, `reward_imp`
- **Session Metrics**: `retention`, `sessions`, `session_length`, `daily_usage`

**Cohort Interval** (string, required for time-based metrics):
- Valid values: `0`, `1`, `2`, `3`, `4`, `5`, `6`, `7`, `10`, `14`, `18`, `21`, `24`, `27`, `30`, `45`
- Specifies days post-install to track (e.g., `7` returns day 0-7 metrics)

**Filters**:
- `filter_country`: Two-letter ISO country code
- `filter_package_name`: App package name
- `filter_platform`: Platform (android, ios)
- `filter_application`: Application name

**Sorting**:
- `sort_day`: Sort by installation day (ASC/DESC)
- `sort_installs`: Sort by number of installs (ASC/DESC)

**Pagination**:
- `limit`: Limit number of results
- `offset`: Skip first N results

**Other**:
- `not_zero` (boolean): Exclude rows where all metrics are zero

#### Example Usage

```
Show me the 7-day revenue per install for users who installed last week
```

Claude will use:
```json
{
  "cohort_type": "revenue",
  "start": "2025-11-25",
  "end": "2025-11-28",
  "format": "json",
  "columns": ["day", "installs", "ads_rpi"],
  "cohort_interval": "7"
}
```

This automatically expands to track `ads_rpi_0` through `ads_rpi_7` for each install cohort.

## Project Structure

```
Applovin-Max-MCP/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── capability.go        # Capability abstraction layer
│   └── toolkit.go           # Revenue and cohort tool implementations
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── .gitignore              # Git ignore rules
└── README.md               # This file
```

## Development

### Building

```bash
go build -o bootstrap cmd/main.go
```

### Dependencies

Key dependencies:
- `github.com/mark3labs/mcp-go` v0.43.0 - MCP protocol implementation
- `github.com/google/uuid` - UUID generation
- `github.com/invopop/jsonschema` - JSON schema handling

See `go.mod` for complete dependency list.

### API Endpoints

The server interacts with the following AppLovin Max API endpoints:

- **Revenue Report**: `https://r.applovin.com/maxReport`
- **Cohort (Revenue)**: `https://r.applovin.com/maxCohort`
- **Cohort (Impression)**: `https://r.applovin.com/maxCohort/imp`
- **Cohort (Session)**: `https://r.applovin.com/maxCohort/session`

## Example Queries

Once integrated with Claude, you can ask natural language questions like:

- "How much revenue did we make in the last 7 days?"
- "Show me the top 5 countries by revenue this month"
- "What's our eCPM trend over the past 30 days?"
- "Analyze retention rates for users who installed last week"
- "Compare iOS vs Android revenue performance"
- "Show me rewarded video impressions by application"

## Error Handling

The server provides detailed error messages for:
- Missing required parameters
- Invalid date formats or ranges
- API authentication failures
- HTTP request failures
- Invalid cohort intervals

## Security Notes

**⚠️ This is a POC with basic security only:**

- Store your API key securely (use environment variables, never commit to version control)
- The API key is transmitted via HTTPS query parameters to AppLovin's secure endpoints
- Ensure proper access controls on the `bootstrap` binary and configuration files
- **For production use**: Implement proper secrets management, request authentication, rate limiting, and audit logging

## Production Readiness Checklist

If you want to use this project as a foundation for production, consider implementing:

- [ ] Comprehensive error handling and recovery
- [ ] Request rate limiting and throttling
- [ ] Input validation and sanitization
- [ ] Structured logging (JSON format with log levels)
- [ ] Monitoring and alerting (metrics, health checks)
- [ ] Unit tests and integration tests (aim for >80% coverage)
- [ ] Security audit and penetration testing
- [ ] Secure credential management (HashiCorp Vault, AWS Secrets Manager, etc.)
- [ ] Request timeout handling and circuit breakers
- [ ] API response caching where appropriate
- [ ] Graceful shutdown and cleanup
- [ ] Documentation for deployment and operations
- [ ] CI/CD pipeline with automated testing
- [ ] Performance profiling and optimization
- [ ] Horizontal scalability considerations

## Version

Current version: **1.0.0** (Proof of Concept)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Summary:** You are free to use, modify, and distribute this software for any purpose, including commercial use, with no warranty provided.

## Contributing

Contributions are welcome! This is an open-source POC project. Feel free to:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure any contributions maintain the spirit of simplicity for a POC project.

## Support

For issues related to:
- **This MCP server**: Open an issue in this repository
- **AppLovin Max API**: Contact AppLovin support
- **Model Context Protocol**: Visit [modelcontextprotocol.io](https://modelcontextprotocol.io)
