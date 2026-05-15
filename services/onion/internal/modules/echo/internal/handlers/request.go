package handlers

// StreamRequest is the JSON body the dash/web client posts to the SSE
// endpoint. New fields (mode/model) are optional so older clients keep
// working unchanged.
type StreamRequest struct {
	Prompt        string `json:"prompt"`
	SessionID     string `json:"session_id"`
	UserID        int64  `json:"user_id"`
	SystemContext string `json:"system_context"`
	Mode          string `json:"mode,omitempty"`
	Model         string `json:"model,omitempty"`
}
