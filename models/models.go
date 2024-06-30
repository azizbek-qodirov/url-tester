package models

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
