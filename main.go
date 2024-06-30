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

	"github.com/Azizbek-Qodirov/url-tester/models"
)

func PerformLoadTests(reqModels []*models.RequestModel) []models.TestResult {
	var results []models.TestResult

	for _, reqModel := range reqModels {
		successfulRequests, failedRequests, duration, logs := performSingleLoadTest(reqModel)
		result := models.TestResult{
			Method:             reqModel.Method,
			URL:                reqModel.URL,
			SuccessfulRequests: successfulRequests,
			FailedRequests:     failedRequests,
			Time:               duration.Seconds(),
			Logs:               string(logs),
		}
		results = append(results, result)
	}
	return results
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

func performSingleLoadTest(reqModel *models.RequestModel) (int, int, time.Duration, []byte) {
	var logs []byte
	var wg sync.WaitGroup
	ch := make(chan int, reqModel.ReqCount)
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()

	if !isValidURL(reqModel.URL) {
		logs = append(logs, []byte(fmt.Sprintf("Invalid URL: %s<br>", reqModel.URL))...)
		return 0, reqModel.ReqCount, time.Since(start), logs
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
	return successfulRequests, failedRequests, dur, logs
}
