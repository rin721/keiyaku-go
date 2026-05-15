package types

type Response struct {
	Code Code        `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func NewResponse(code Code, msg string, data interface{}) Response {
	if msg == "" {
		msg = Message(code)
	}
	return Response{Code: code, Msg: msg, Data: data}
}

func OK(data interface{}) Response {
	return NewResponse(CodeOK, MessageOK, data)
}

func EmptyOK() Response {
	return NewResponse(CodeOK, MessageOK, nil)
}
