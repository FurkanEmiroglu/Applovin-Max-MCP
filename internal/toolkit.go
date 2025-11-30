package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

var allowed_ranges []int = []int{0, 1, 2, 3, 4, 5, 6, 7, 10, 14, 18, 21, 24, 27, 30, 45}

func GetApiKey() string {
	return os.Getenv("APPLOVIN_API_KEY")
}

func NewCohortRequestCapability() Capability {
	return Capability{
		Tool: mcp.NewTool("cohort_request",
			mcp.WithDescription(
				"sends a cohort request to AppLovin Max MCP",
			),
			mcp.WithString("sort_day",
				mcp.Description("Sort type"),
				mcp.Enum("ASC", "DESC"),
				mcp.Required(),
			),
			mcp.WithString("start",
				mcp.Description("YYYY-MM-DD formatted starting date."),
			),
			mcp.WithString("end",
				mcp.Description("YYYY-MM-DD formatted ending date."),
			),
			mcp.WithString("format",
				mcp.Description("output format"),
				mcp.Enum("json", "csv"),
			),
			mcp.WithArray("columns",
				mcp.Description("which columns to report"),
				mcp.DefaultArray([]string{"day", "installs"}),
				mcp.WithStringEnumItems([]string{"day", "installs", "country", "inter_rpi", "banner_rpi", "reward_rpi", "all_rpi", "platform", "package_name"}),
			),
			mcp.WithString("filter_country",
				mcp.Description("country filter, two letter iso code"),
			),
			mcp.WithString("filter_package_name",
				mcp.Description("package_name filter"),
			),
			mcp.WithString("filter_platform",
				mcp.Description("platform filter"),
			),
		),
		Handler: func(ctx context.Context, toolRequest mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			queryParameters := url.Values{}
			queryParameters.Add("api_key", GetApiKey())
			queryParameters.Add("sort_day", "ASC")

			addQueryParametersIfGiven := func(args map[string]any, objectNames ...string) {
				for _, objName := range objectNames {
					if objValue, ok := args[objName].(string); ok {
						queryParameters.Add(objName, strings.ToLower(objValue))
					}
				}
			}

			args := toolRequest.GetArguments()

			addQueryParametersIfGiven(args, "start", "end", "format")

			if columns, ok := args["columns"].([]any); ok {
				columnList := make([]string, len(columns))

				for i, v := range columns {
					str, ok := v.(string)
					if !ok {
						continue
					}

					if strings.Contains(str, "_rpi") {
						continue
					}

					columnList[i] = str
				}

				queryParameters.Add("columns", strings.ToLower(strings.Join(columnList, ",")))
			}

			checkFilterAndApply := func(filtersObject map[string]any, filterNames ...string) {
				filterPrefix := "filter_"

				for _, filterName := range filterNames {
					if filterValue, ok := filtersObject[filterName].(string); ok {
						queryParameters.Add(fmt.Sprintf("%v%v", filterPrefix, filterName), filterValue)
					}
				}
			}

			if filters, ok := args["filters"].(map[string]any); ok {
				checkFilterAndApply(filters, "package_name", "platform", "country")
			}

			url, err := url.Parse("https://r.applovin.com/maxCohort")

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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return mcp.NewToolResultErrorf("error while reading max report body: %v", err), nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				return mcp.NewToolResultErrorf("max api returned status code: %v. request body: %v", resp.StatusCode, string(body)), nil
			}

			return mcp.NewToolResultText(string(body)), nil
		},
	}
}
