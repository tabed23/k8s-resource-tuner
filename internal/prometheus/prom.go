package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type PromClient struct {
	BaseURL string
	Client  *http.Client
}

type promQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func NewPromClient(url string) *PromClient {
	return &PromClient{
		BaseURL: url,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}
func (pc *PromClient) QueryRange(query string, start, end time.Time, step string) ([]float64, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/query_range", pc.BaseURL))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", start.Unix()))
	q.Set("end", fmt.Sprintf("%d", end.Unix()))
	q.Set("step", step)
	u.RawQuery = q.Encode()

	resp, err := pc.Client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result promQueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", result.Status)
	}

	values := []float64{}
	for _, res := range result.Data.Result {
		for _, v := range res.Values {
			if len(v) < 2 {
				continue
			}
			valStr, ok := v[1].(string)
			if !ok {
				continue
			}
			var f float64
			fmt.Sscanf(valStr, "%f", &f)
			values = append(values, f)
		}
	}
	return values, nil
}

func (pc *PromClient) QueryCpu(namespace string, deploy string, start, end time.Time, step string) ([]float64, error) {
	query := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s", pod=~"%s-.*"}[5m])) by (pod)`, namespace, deploy)
	return pc.QueryRange(query, start, end, step)

}
func (pc *PromClient) QueryMemory(namespace string, deploy string, start, end time.Time, step string) ([]float64, error) {
	query := fmt.Sprintf(`max_over_time(container_memory_usage_bytes{namespace="%s", pod=~"%s-.*"}[5m])`, namespace, deploy)
	return pc.QueryRange(query, start, end, step)
}
