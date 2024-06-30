package models

type RequestModel struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Body      string            `json:"body"`
	Headers   map[string]string `json:"headers"`
	ReqCount  int               `json:"req_count"`
	CReqCount int               `json:"c_req_count"`
}

type TestResult struct {
	Method             string  `json:"method"`
	URL                string  `json:"url"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	Time               float64 `json:"time"` // Duration in seconds
	Logs               string  `json:"logs"`
}
