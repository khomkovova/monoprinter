package models

type Response struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`
	Data              string `json:"data"`
}
