package apis

const (
	_                = iota
	Ok               // Response code for success
	Err              // Response code for error
	InvalidArguments // Response code for invalid arguments
)

type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ResponseOk() BaseResponse {
	return BaseResponse{
		Code:    Ok,
		Message: "ok",
	}
}

func ResponseOkWithMsg(msg string) BaseResponse {
	return BaseResponse{
		Code:    Ok,
		Message: msg,
	}
}

type ListEngineResponse struct {
	BaseResponse
	Engines []EngineModel `json:"engines"`
}
