package url_tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type RequestModel struct {
	URL       string            `json:"url"`         // URL
	Method    string            `json:"method"`      // Method
	Body      string            `json:"body"`        // Body
	Headers   map[string]string `json:"headers"`     // Headers
	ReqCount  int               `json:"req_count"`   // Request count
	CReqCount int               `json:"c_req_count"` // Concurrent request counts
}

type TestResult struct {
	Method             string  `json:"method"`              // Method
	URL                string  `json:"url"`                 // URL
	SuccessfulRequests int     `json:"successful_requests"` // Succesful Requests
	FailedRequests     int     `json:"failed_requests"`     // Failed Requests
	Time               float64 `json:"time"`                // Duration in seconds
	Logs               []byte  `json:"logs"`                // Logs
}

// InitRequest initializes and returns a pointer to a new RequestModel instance.
// This function is useful for creating a new request model with default values.
//
// Returns a struct:
//
//	type RequestModel struct {
//		URL       string            `json:"url"`
//		Method    string            `json:"method"`
//		Body      string            `json:"body"`
//		Headers   map[string]string `json:"headers"`
//		ReqCount  int               `json:"req_count"`
//		CReqCount int               `json:"c_req_count"`
//	}
func InitRequest() *RequestModel {
	return &RequestModel{}
}

// InitRequests initializes and returns a slice of pointers to RequestModel instances.
// This function is useful for creating an empty slice to hold multiple request models.
//
// Returns a slice to store:
//
//	type RequestModel struct {
//		URL       string            `json:"url"`
//		Method    string            `json:"method"`
//		Body      string            `json:"body"`
//		Headers   map[string]string `json:"headers"`
//		ReqCount  int               `json:"req_count"`
//		CReqCount int               `json:"c_req_count"`
//	}
func InitRequests() []*RequestModel {
	return []*RequestModel{}
}

// DoTest performs load tests for multiple request models.
// It takes a slice of RequestModel pointers and returns a slice of TestResult.
// Each TestResult contains details about the successful and failed requests,
// the duration of the test, and logs for each request.
func DoTests(reqModels []*RequestModel) []TestResult {
	var results []TestResult

	for _, reqModel := range reqModels {
		result := performSingleLoadTest(reqModel)
		results = append(results, result)
	}
	return results
}

// DoTests performs a load test for a single request model.
// It takes a pointer to a RequestModel and returns a TestResult.
// The TestResult contains details about the successful and failed requests,
// the duration of the test, and logs for each request.
func DoTest(reqModel *RequestModel) TestResult {
	return performSingleLoadTest(reqModel)
}

func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	if u.Host == "localhost" || strings.HasPrefix(u.Host, "localhost:") {
		return true
	}

	parts := strings.Split(u.Host, ".")
	if len(parts) < 2 {
		return false
	}
	tld := parts[len(parts)-1]
	if len(tld) < 2 || len(tld) > 6 {
		return false
	}

	_, err = net.LookupHost(u.Host)
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Head(urlStr)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400
}

func performSingleLoadTest(reqModel *RequestModel) TestResult {
	var logs []byte
	var wg sync.WaitGroup
	ch := make(chan int, reqModel.ReqCount)
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()

	if !isValidURL(reqModel.URL) {
		logs = append(logs, []byte(fmt.Sprintf("Invalid URL: %s<br>", reqModel.URL))...)
		return TestResult{
			Method:             reqModel.Method,
			URL:                reqModel.URL,
			SuccessfulRequests: 0,
			FailedRequests:     reqModel.ReqCount,
			Time:               time.Since(start).Seconds(),
			Logs:               logs,
		}
	}

	for i := 0; i < reqModel.ReqCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err := http.NewRequest(reqModel.Method, reqModel.URL, bytes.NewBuffer([]byte(reqModel.Body)))
			if err != nil {
				logs = append(logs, []byte(fmt.Sprintf("%s &emsp;&emsp; %d &emsp;&emsp; %s &emsp;&emsp;Error creating request: nil <br>", reqModel.Method, 0, reqModel.URL))...)
				ch <- 0
				return
			}

			for key, value := range reqModel.Headers {
				req.Header.Set(key, value)
			}

			resp, err := client.Do(req)
			if err != nil {
				if resp != nil {
					logs = append(logs, []byte(fmt.Sprintf("%s &emsp;&emsp; %d &emsp;&emsp; %s &emsp;&emsp;Error performing request: %s <br>", reqModel.Method, resp.StatusCode, reqModel.URL, err.Error()))...)
				} else {
					logs = append(logs, []byte(fmt.Sprintln("Goroutine error: Invalid memory address or nil pointer dereference"))...)
				}
				ch <- 0
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				res := fmt.Sprintf("%s &emsp;&emsp; %d &emsp;&emsp; %s &emsp;&emsp;Error: nil <br>", reqModel.Method, resp.StatusCode, reqModel.URL)
				logs = append(logs, []byte(res)...)
				ch <- 1
			} else {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					logs = append(logs, []byte(fmt.Sprintf("%s &emsp;&emsp; %d &emsp;&emsp; %s &emsp;&emsp;Error reading response body: %s <br>", reqModel.Method, resp.StatusCode, reqModel.URL, err.Error()))...)
					ch <- 0
					return
				}
				type ErrResp struct {
					Error   string `json:"error"`
					Details string `json:"details"`
				}
				var errResp ErrResp
				err = json.Unmarshal(body, &errResp)
				if err != nil {
					logs = append(logs, []byte(fmt.Sprintf("%s &emsp;&emsp; %d &emsp;&emsp; %s &emsp;&emsp;Error unmarshalling response: %s <br>", reqModel.Method, resp.StatusCode, reqModel.URL, err.Error()))...)
					ch <- 0
					return
				}
				bodyText := errResp.Error + " " + errResp.Details
				res := fmt.Sprintf("%s &emsp;&emsp; %d &emsp;&emsp; %s &emsp;&emsp;Error: %s <br>", reqModel.Method, resp.StatusCode, reqModel.URL, bodyText)
				logs = append(logs, []byte(res)...)
				ch <- 0
			}
		}()

		if (i+1)%reqModel.CReqCount == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
	close(ch)

	successfulRequests := 0
	failedRequests := 0
	for status := range ch {
		if status == 1 {
			successfulRequests++
		} else {
			failedRequests++
		}
	}
	dur := time.Since(start)
	return TestResult{
		Method:             reqModel.Method,
		URL:                reqModel.URL,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		Time:               dur.Seconds(),
		Logs:               logs,
	}
}
