package web

type TaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	// Status  string `json:"status"`
}
