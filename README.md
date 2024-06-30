# URL Tester

URL Tester is a Go package for load testing APIs. It provides functionalities to perform load tests with various HTTP methods and request configurations.

## Installation

To install URL Tester, use `go get`:

```sh
go get github.com/Azizbek-Qodirov/url-tester
```

# Usage
## Initialize a Request Model
To initialize a single request model with default values:

```go
reqModel := url_tester.InitRequest()
```

## Initialize Multiple Request Models
To initialize a slice of request models for batch testing:

```go
reqModels := url_tester.InitRequests()
```

## Perform Load Tests
To perform load tests using a slice of request models:

```go
results := url_tester.DoTests(reqModels)
```

## Single Test Execution
To perform a load test for a single request model:

```go
result := url_tester.DoTest(reqModel)
```

# Example
Here's a simple example demonstrating how to perform a load test:
```go
package main

import (
	"fmt"
	"github.com/Azizbek-Qodirov/url-tester"
)

func main() {
	reqModel := url_tester.InitRequest()
	reqModel.URL = "https://example.com"
	reqModel.Method = "GET"
	reqModel.ReqCount = 10
	reqModel.CReqCount = 5

	result := url_tester.DoTest(reqModel)

	fmt.Printf("Test Result: %+v\n", result)
}
```