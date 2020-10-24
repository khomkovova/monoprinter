package models

type Response struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`
	Data              string `json:"data"`
	StatusCode        string `json:"status_code"`
}

type GCPFile struct {
	FileUrl string `json:"file_url"`
}