package customresponse

import "encoding/json"

const ERROR_STATUS = "error"
const ERROR_COOKIES  = "error_cookies"
const ERROR_SERVER = "error_server"
const ERROR_REQUEST = "error_request"
const ERROR_USER  =  "error_user"
const OK_STATUS  = "ok"


type customResponse struct {
	Status            string `json:"status"`
	StatusCode        string `json:"status_code"`
	StatusDescription string `json:"status_description"`
	Data              string `json:"data"`
}

func New(status string, statusCode string, statusDescription string, data string) customResponse  {
	return customResponse{status, statusCode, statusDescription, data}
}

func (response *customResponse) GetByteResponse()(byteResponse []byte ) {
	byteResponse, _ = json.Marshal(response)
	return byteResponse
}
