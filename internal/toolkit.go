package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ApplovinToolkit struct {
	ApiKey string
}

func (a *ApplovinToolkit) Setup(serv *server.MCPServer) {
	serv.AddTool(a.CohortRequest())
}

func NewAgent(apiKey string) *ApplovinToolkit {
	return &ApplovinToolkit{
		ApiKey: apiKey,
	}
}

func (a *ApplovinToolkit) CohortRequest() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool(
			"cohort_request",
			mcp.WithDescription("cohort_Request_description_here"),
		),
		func(ctx context.Context, callReq mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			queryParameters := url.Values{}
			queryParameters.Add("api_key", a.ApiKey)
			queryParameters.Add("sort_day", "ASC")

			addQueryParametersIfGiven := func(args map[string]interface{}, objectNames ...string) {
				for _, objName := range objectNames {
					if objValue, ok := args[objName].(string); ok {
						queryParameters.Add(objName, strings.ToLower(objValue))
					}
				}
			}

			args := callReq.GetArguments()

			addQueryParametersIfGiven(args, "start", "end", "format")

			if breakdowns, ok := args["breakdowns"].([]any); ok {
				breakdownList := make([]string, len(breakdowns))

				for i, v := range breakdowns {
					str, ok := v.(string)
					if !ok {
						continue
					}

					breakdownList[i] = str
				}

				queryParameters.Add("columns", strings.ToLower(strings.Join(breakdownList, ",")))
			}

			checkFilterAndApply := func(filtersObject map[string]any, filterNames ...string) {
				filterPrefix := "filter_"

				for _, filterName := range filterNames {
					if filterValue, ok := filtersObject[filterName].(string); ok {
						queryParameters.Add(fmt.Sprintf("%v%v", filterPrefix, filterName), filterValue)
					}
				}
			}

			if filters, ok := args["filters"].(map[string]interface{}); ok {
				checkFilterAndApply(filters, "package_name", "platform", "country")
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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return mcp.NewToolResultErrorf("error while reading max report body: %v", err), nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				return mcp.NewToolResultErrorf("max api returned status code: %v. request body: %v", resp.StatusCode, string(body)), nil
			}

			return mcp.NewToolResultText(string(body)), nil
		}
}
