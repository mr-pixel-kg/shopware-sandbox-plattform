package dto

const (
	TerminalMsgResize = "resize"
	TerminalMsgExit   = "exit"
	TerminalMsgError  = "error"
)

type TerminalControlMessage struct {
	Type    string `json:"type"`
	Cols    uint   `json:"cols,omitempty"`
	Rows    uint   `json:"rows,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
