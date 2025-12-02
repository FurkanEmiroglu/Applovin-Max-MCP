package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

var allowedRanges []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "10", "14", "18", "21", "24", "27", "30", "45"}

func GetApiKey() string {
	return os.Getenv("APPLOVIN_API_KEY")
}

func NewRevenueReportCapability() Capability {
	return Capability{
		Tool: mcp.NewTool("revenue_report",
			mcp.WithDescription(
				"sends a revenue report request to AppLovin Max API for aggregated mediation statistics",
			),
			mcp.WithString("start",
				mcp.Description("YYYY-MM-DD formatted starting date."),
				mcp.Required(),
			),
			mcp.WithString("end",
				mcp.Description("YYYY-MM-DD formatted ending date."),
				mcp.Required(),
			),
			mcp.WithString("format",
				mcp.Description("output format"),
				mcp.Enum("json", "csv"),
				mcp.Required(),
			),
			mcp.WithArray("columns",
				mcp.Description("Metrics to include in the report. Use dimension columns to group data (day, hour, application, country, platform) and metric columns to get performance data (impressions, responses, estimated_revenue, ecpm, fill_rate). You can combine multiple dimensions and metrics in a single request."),
				mcp.DefaultArray([]string{"day", "application"}),
				mcp.WithStringEnumItems([]string{
					"day", "hour", "application", "package_name", "ad_format", "country",
					"platform", "network", "network_placement", "max_placement", "max_ad_unit_id",
					"custom_network_name", "ad_unit_waterfall_name", "device_type", "store_id",
					"has_idfa", "max_ad_unit_test",
					"impressions", "responses", "requests", "attempts", "estimated_revenue",
					"ecpm", "fill_rate",
				}),
			),
			mcp.WithString("filter_application",
				mcp.Description("application filter"),
			),
			mcp.WithString("filter_package_name",
				mcp.Description("package_name filter"),
			),
			mcp.WithString("filter_ad_type",
				mcp.Description("ad_type filter (e.g., banner, inter, rewarded)"),
			),
			mcp.WithString("filter_country",
				mcp.Description("country filter, two letter iso code"),
			),
			mcp.WithString("filter_platform",
				mcp.Description("platform filter (e.g., android, ios)"),
			),
			mcp.WithString("filter_network",
				mcp.Description("network filter"),
			),
			mcp.WithString("filter_zone",
				mcp.Description("zone filter"),
			),
			mcp.WithString("sort_day",
				mcp.Description("Sort by day"),
				mcp.Enum("ASC", "DESC"),
			),
			mcp.WithString("sort_hour",
				mcp.Description("Sort by hour"),
				mcp.Enum("ASC", "DESC"),
			),
			mcp.WithString("sort_estimated_revenue",
				mcp.Description("Sort by estimated revenue"),
				mcp.Enum("ASC", "DESC"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Limit number of results for pagination"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Offset for pagination"),
			),
			mcp.WithBoolean("not_zero",
				mcp.Description("Exclude results where all numerical metrics equal zero"),
			),
		),
		Handler: func(ctx context.Context, toolRequest mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			queryParameters := url.Values{}
			queryParameters.Add("api_key", GetApiKey())

			checkParameter := func(args map[string]any, key string) (*string, bool) {
				for k, v := range args {
					if key != k {
						continue
					}

					if valueStr, ok := v.(string); ok {
						return &valueStr, true
					}
				}

				return nil, false
			}

			args := toolRequest.GetArguments()

			if startDate, ok := checkParameter(args, "start"); ok {
				queryParameters.Add("start", strings.ToLower(*startDate))
			} else {
				return mcp.NewToolResultError("start date required"), nil
			}

			if endDate, ok := checkParameter(args, "end"); ok {
				queryParameters.Add("end", strings.ToLower(*endDate))
			} else {
				return mcp.NewToolResultError("end date required"), nil
			}

			if format, ok := checkParameter(args, "format"); ok {
				queryParameters.Add("format", strings.ToLower(*format))
			} else {
				return mcp.NewToolResultError("format required"), nil
			}

			if columns, ok := args["columns"].([]any); ok {
				var columnsStr []string = []string{}
				for _, v := range columns {
					if v, ok := v.(string); ok {
						columnsStr = append(columnsStr, v)
					}
				}
				queryParameters.Add("columns", strings.ToLower(strings.Join(columnsStr, ",")))
			} else {
				return mcp.NewToolResultError("columns required"), nil
			}

			// Add filters
			if filter, ok := checkParameter(args, "filter_application"); ok {
				queryParameters.Add("filter_application", *filter)
			}

			if filter, ok := checkParameter(args, "filter_package_name"); ok {
				queryParameters.Add("filter_package_name", *filter)
			}

			if filter, ok := checkParameter(args, "filter_ad_type"); ok {
				queryParameters.Add("filter_ad_type", *filter)
			}

			if filter, ok := checkParameter(args, "filter_country"); ok {
				queryParameters.Add("filter_country", *filter)
			}

			if filter, ok := checkParameter(args, "filter_platform"); ok {
				queryParameters.Add("filter_platform", *filter)
			}

			if filter, ok := checkParameter(args, "filter_network"); ok {
				queryParameters.Add("filter_network", *filter)
			}

			if filter, ok := checkParameter(args, "filter_zone"); ok {
				queryParameters.Add("filter_zone", *filter)
			}

			// Add sorting
			if sort, ok := checkParameter(args, "sort_day"); ok {
				queryParameters.Add("sort_day", *sort)
			}

			if sort, ok := checkParameter(args, "sort_hour"); ok {
				queryParameters.Add("sort_hour", *sort)
			}

			if sort, ok := checkParameter(args, "sort_estimated_revenue"); ok {
				queryParameters.Add("sort_estimated_revenue", *sort)
			}

			// Add pagination
			if limit, ok := args["limit"].(float64); ok {
				queryParameters.Add("limit", strconv.Itoa(int(limit)))
			}

			if offset, ok := args["offset"].(float64); ok {
				queryParameters.Add("offset", strconv.Itoa(int(offset)))
			}

			// Add not_zero
			if notZero, ok := args["not_zero"].(bool); ok {
				if notZero {
					queryParameters.Add("not_zero", "1")
				}
			}

			url, err := url.Parse("https://r.applovin.com/maxReport")

			if err != nil {
				return mcp.NewToolResultErrorf("error while parsing max report url: %v", err), nil
			}

			url.RawQuery = queryParameters.Encode()

			req, err := http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return nil, err
			}

			req.Header.Add("Accept", "application/json")
			client := http.DefaultClient
			resp, err := client.Do(req)

			if err != nil {
				return mcp.NewToolResultErrorf("error while sending max report request: %v", err), nil
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return mcp.NewToolResultErrorf("error while reading max report body: %v", err), nil
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				return mcp.NewToolResultErrorf("max api returned status code: %v. request body: %v", resp.StatusCode, string(body)), nil
			}

			return mcp.NewToolResultText(string(body)), nil
		},
	}
}

func NewCohortRequestCapability() Capability {
	return Capability{
		Tool: mcp.NewTool("cohort_request",
			mcp.WithDescription(
				"retrieves user cohort performance data segmented by installation date from AppLovin Max. Use this to analyze how revenue, impressions, or sessions evolve over time for users who installed on specific dates. Choose the cohort_type based on what metrics you need: 'revenue' for ad revenue and IAP, 'impression' for ad impressions, or 'session' for retention and usage patterns.",
			),
			mcp.WithString("cohort_type",
				mcp.Description("Type of cohort data to retrieve: 'revenue' for ad revenue/IAP metrics (default), 'impression' for ad impression data, 'session' for retention/session metrics"),
				mcp.Enum("revenue", "impression", "session"),
				mcp.DefaultString("revenue"),
			),
			mcp.WithString("start",
				mcp.Description("Start date for the cohort analysis in YYYY-MM-DD format. This is the installation date to start from."),
				mcp.Required(),
			),
			mcp.WithString("end",
				mcp.Description("End date for the cohort analysis in YYYY-MM-DD format. Maximum 45-day range from start date."),
				mcp.Required(),
			),
			mcp.WithString("format",
				mcp.Description("Response format: 'json' for structured data or 'csv' for comma-separated values"),
				mcp.Enum("json", "csv"),
				mcp.Required(),
			),
			mcp.WithArray("columns",
				mcp.Description("Metrics to include in the report. IMPORTANT: When requesting revenue/impression/session metrics with day suffixes (e.g., _rpi_X, _imp_X, _retention_X), you MUST also specify cohort_interval. Common dimension columns: day (install date), installs, country, platform, package_name, application. Common revenue metrics: ads_rpi (ad revenue per install), iap_rpi (in-app purchase revenue per install), pub_revenue (publisher revenue), inter_rpi/banner_rpi/reward_rpi (by ad type). Common impression metrics: imp (impressions), imp_per_user, banner_imp, inter_imp, reward_imp. Common session metrics: retention (% users active), sessions, session_length, daily_usage."),
				mcp.DefaultArray([]string{"day", "installs"}),
				mcp.WithStringEnumItems([]string{
					"day", "installs", "country", "platform", "package_name", "application",
					"ads_rpi", "iap_rpi", "pub_revenue", "inter_rpi", "banner_rpi", "reward_rpi",
					"imp", "imp_per_user", "banner_imp", "inter_imp", "reward_imp",
					"retention", "sessions", "session_length", "daily_usage",
				}),
			),
			mcp.WithString(
				"cohort_interval",
				mcp.Description("REQUIRED when using time-based metrics (columns with _rpi, _imp, _retention suffixes). Specifies how many days post-install to track. For example, cohort_interval=7 will return metrics for days 0-6 after install (e.g., ads_rpi_0, ads_rpi_1, ... ads_rpi_6). Use 0 for install day only, 7 for first week, 30 for first month, 45 for maximum range."),
				mcp.Enum(allowedRanges...),
			),
			mcp.WithString("filter_country",
				mcp.Description("Filter results to a specific country using two-letter ISO code (e.g., 'US', 'GB', 'JP')"),
			),
			mcp.WithString("filter_package_name",
				mcp.Description("Filter results to a specific app package name (e.g., 'com.example.app')"),
			),
			mcp.WithString("filter_platform",
				mcp.Description("Filter results to a specific platform: 'android' or 'ios'"),
			),
			mcp.WithString("filter_application",
				mcp.Description("Filter results to a specific application name"),
			),
			mcp.WithString("sort_day",
				mcp.Description("Sort results by installation day in ascending (ASC) or descending (DESC) order"),
				mcp.Enum("ASC", "DESC"),
			),
			mcp.WithString("sort_installs",
				mcp.Description("Sort results by number of installs in ascending (ASC) or descending (DESC) order"),
				mcp.Enum("ASC", "DESC"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Limit the number of results returned (for pagination)"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Skip the first N results (for pagination, use with limit)"),
			),
			mcp.WithBoolean("not_zero",
				mcp.Description("When true, exclude rows where all metric values are zero"),
			),
		),
		Handler: func(ctx context.Context, toolRequest mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			queryParameters := url.Values{}
			queryParameters.Add("api_key", GetApiKey())

			checkParameter := func(args map[string]any, key string) (*string, bool) {
				for k, v := range args {
					if key != k {
						continue
					}

					if valueStr, ok := v.(string); ok {
						return &valueStr, true
					}
				}

				return nil, false
			}

			args := toolRequest.GetArguments()

			// Determine cohort endpoint based on type
			cohortEndpoint := "https://r.applovin.com/maxCohort"
			if cohortType, ok := checkParameter(args, "cohort_type"); ok {
				switch *cohortType {
				case "impression":
					cohortEndpoint = "https://r.applovin.com/maxCohort/imp"
				case "session":
					cohortEndpoint = "https://r.applovin.com/maxCohort/session"
				case "revenue":
					cohortEndpoint = "https://r.applovin.com/maxCohort"
				}
			}

			if startDate, ok := checkParameter(args, "start"); ok {
				queryParameters.Add("start", strings.ToLower(*startDate))
			} else {
				return mcp.NewToolResultError("start date required"), nil
			}

			if endDate, ok := checkParameter(args, "end"); ok {
				queryParameters.Add("end", strings.ToLower(*endDate))
			} else {
				return mcp.NewToolResultError("end date required"), nil
			}

			if format, ok := checkParameter(args, "format"); ok {
				queryParameters.Add("format", strings.ToLower(*format))
			} else {
				return mcp.NewToolResultError("format required"), nil
			}

			if columns, ok := args["columns"].([]any); ok {
				var columnsStr []string = []string{}

				for _, v := range columns {
					if v, ok := v.(string); ok {
						// Check if this is a time-based metric column
						if strings.HasSuffix(v, "_rpi") || strings.HasSuffix(v, "_imp") ||
							strings.HasSuffix(v, "_retention") || strings.Contains(v, "_per_user") ||
							v == "pub_revenue" || v == "sessions" || v == "session_length" || v == "daily_usage" {
							interval, ok := checkParameter(args, "cohort_interval")
							if !ok {
								return mcp.NewToolResultError(fmt.Sprintf("cohort_interval required when using time-based metric: %s", v)), nil
							}

							intervalInt, err := strconv.Atoi(*interval)
							if err != nil {
								return mcp.NewToolResultError("cohort_interval string couldn't be parsed to int"), nil
							}

							if !strings.Contains(strings.Join(allowedRanges, ","), *interval) {
								return mcp.NewToolResultError(fmt.Sprintf("invalid cohort_interval: %s", *interval)), nil
							}

							// Generate columns for each day in the interval (0 to interval-1)
							for i := 0; i <= intervalInt; i++ {
								columnsStr = append(columnsStr, fmt.Sprintf("%s_%d", v, i))
							}
							continue
						}
						columnsStr = append(columnsStr, v)
					}
				}
				queryParameters.Add("columns", strings.ToLower(strings.Join(columnsStr, ",")))
			} else {
				return mcp.NewToolResultError("columns required"), nil
			}

			// Add filters
			if filter, ok := checkParameter(args, "filter_country"); ok {
				queryParameters.Add("filter_country", *filter)
			}

			if filter, ok := checkParameter(args, "filter_package_name"); ok {
				queryParameters.Add("filter_package_name", *filter)
			}

			if filter, ok := checkParameter(args, "filter_platform"); ok {
				queryParameters.Add("filter_platform", *filter)
			}

			if filter, ok := checkParameter(args, "filter_application"); ok {
				queryParameters.Add("filter_application", *filter)
			}

			// Add sorting
			if sort, ok := checkParameter(args, "sort_day"); ok {
				queryParameters.Add("sort_day", *sort)
			}

			if sort, ok := checkParameter(args, "sort_installs"); ok {
				queryParameters.Add("sort_installs", *sort)
			}

			// Add pagination
			if limit, ok := args["limit"].(float64); ok {
				queryParameters.Add("limit", strconv.Itoa(int(limit)))
			}

			if offset, ok := args["offset"].(float64); ok {
				queryParameters.Add("offset", strconv.Itoa(int(offset)))
			}

			// Add not_zero
			if notZero, ok := args["not_zero"].(bool); ok {
				if notZero {
					queryParameters.Add("not_zero", "1")
				}
			}

			url, err := url.Parse(cohortEndpoint)

			if err != nil {
				return mcp.NewToolResultErrorf("error while parsing max report url: %v", err), nil
			}

			url.RawQuery = queryParameters.Encode()

			req, err := http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return nil, err
			}

			req.Header.Add("Accept", "application/json")
			client := http.DefaultClient
			resp, err := client.Do(req)

			if err != nil {
				return mcp.NewToolResultErrorf("error while sending max report request: %v", err), nil
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return mcp.NewToolResultErrorf("error while reading max report body: %v", err), nil
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				return mcp.NewToolResultErrorf("max api returned status code: %v. request body: %v", resp.StatusCode, string(body)), nil
			}

			return mcp.NewToolResultText(string(body)), nil
		},
	}
}
