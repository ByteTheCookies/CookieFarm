package protocols

// ResponseProtocol represents a response from the flag checker service.
type ResponseProtocol struct {
	Status string `json:"status"` // Status of the response (e.g., "success", "error")
	Flag   string `json:"flag"`   // Flag string received from the flag checker service
	Msg    string `json:"msg"`    // Message from the flag checker service
}
