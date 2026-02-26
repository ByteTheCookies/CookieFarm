# 📡 Communication Protocol Guide

This document describes the communication protocol used to interact with the **Flag Checker** service. It outlines the structure for submitting flags, the response format, and the function signature to communicate with the service.

---

### 🔧 File Structure and Template

Create the file in the `server/protocols` directory and use the following template as a starting point:

```go
//go:build ignore

package main

import (
	"github.com/ByteTheCookies/cookieserver/internal/models"
)

// Submit function sends flags to the Flag Checker service and returns a response.
func Submit(host string, team_token string, flags []string) ([]protocols.ResponseProtocol, error) {
	// Your implementation here
	return nil, nil
}
```

In this example:
- The `Submit` function is designed to send flags to the Flag Checker.
- The function should return a slice of `ResponseProtocol` objects and any errors encountered.

### 📝 `models.ResponseProtocol` Structure

The `ResponseProtocol` struct is used to define the structure of the responses returned by the Flag Checker service:

```go
type ResponseProtocol struct {
	Msg    string `json:"msg"`    // A message containing the flag and additional context.
	Flag   string `json:"flag"`   // The flag that was submitted.
	Status string `json:"status"` // The status of the flag submission.
}
```

### 🔁 Response Format

The Flag Checker will return a **JSON array** of objects with the following structure for each flag submission:

```json
{
  "msg": "[<flag>] <message>",
  "flag": "<flag>",
  "status": "ACCEPTED" | "DENIED" | "RESUBMIT" | "ERROR"
}
```

Where:
- `msg`: A human-readable message that includes the flag and additional context or result.
- `flag`: The original flag that was submitted.
- `status`: The result of the submission, which can be one of the following:
  - `"ACCEPTED"` – The flag is valid and was successfully accepted.
  - `"DENIED"` – The flag is invalid, incorrect, or already submitted.
  - `"RESUBMIT"` – Temporary failure; the client should retry later.
  - `"ERROR"` – A generic error occurred during processing.

### 📤 Submit Function

The `Submit` function should be used to interact with the Flag Checker service. The function signature is as follows:

```go
func Submit(host string, token string, flags []string) ([]protocols.ResponseProtocol, error)
```

Parameters:
- `host`: The address of the Flag Checker service (e.g., `http://localhost:5000`).
- `token`: The authentication token for the team, required for submitting flags.
- `flags`: A slice of strings containing the flags to be submitted.

The function returns:
- A slice of `models.ResponseProtocol` that contains the result for each flag.
- An error if there is a failure during the request or while processing the response.

---

### 🚀 Example Implementation

Here is an example of how you might implement the `Submit` function using an HTTP request to interact with the Flag Checker service:

```go
//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ByteTheCookies/cookieserver/internal/models"
)

func Submit(host string, team_token string, flags []string) ([]protocols.ResponseProtocol, error) {
	jsonData, err := json.Marshal(flags)
	if err != nil {
		return nil, fmt.Errorf("error during marshalling: %w", err)
	}

	url := "http://" + host + "/submit"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error during request creation: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Team-Token", team_token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during request submission: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error during response reading: %w", err)
	}

	var response []models.ResponseProtocol
	// logger.Debug("Raw body %s", string(body))
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error during response parsing: %w", err)
	}

	return response, nil
}

```

This example shows how to send a POST request to the Flag Checker API, including the necessary flags and authentication token, and how to handle the response.

---

### 🛠 Compilation Instructions

To compile the Go code into a shared object (`.so`), use the following command:

```bash
go build -buildmode=plugin -o <name>.so <name>.go
```

Explanation:
- `-buildmode=plugin`: This flag instructs Go to build a plugin, which can be dynamically loaded at runtime.
- `<name>.so`: The output file name for the shared object.
- `<name>.go`: The Go source file to compile.

Make sure the Go file (`<name>.go`) implements the necessary logic for interacting with the Flag Checker and returning the correct response format.

---

### ⚠️ Error Handling

Ensure to handle the following errors:
- **Network errors**: Connection failure to the Flag Checker service.
- **Response errors**: Incorrect or malformed responses from the Flag Checker.
- **JSON parsing errors**: Ensure the response is correctly parsed into the `ResponseProtocol` struct.

If there is an error at any stage, the `Submit` function should return `nil` for the response and the corresponding error.

---

### 🏁 Final Notes

- Make sure that the Flag Checker service is running and accessible via the provided `host`.
- Use valid flags and ensure the authentication token is correct for the requests to succeed.
- The structure of the response from the Flag Checker should strictly follow the provided format to ensure smooth parsing and handling.
