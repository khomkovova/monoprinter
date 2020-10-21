package helper

import (
	"encoding/json"
	"github.com/khomkovova/MonoPrinter/models"
	"log"
    "runtime"
)

func GenerateErrorMsg(err error, comment string) ([]byte, error) {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	log.Printf("Error in function: %s\n", details.Name())
	log.Printf("With error msg: %s\n", err)
	log.Printf("Comment: %s\n", comment)
	var response models.Response
	response.Status = "error"
	response.StatusDescription = comment
	responseByte, _ := json.Marshal(response)
	return responseByte, nil
}

func GenerateOkMsg(data string, comment string) ([]byte, error) {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	log.Printf("Info msg from function: %s\n", details.Name())
	log.Printf("Comment: %s\n", comment)
	var response models.Response
	response.Status = "ok"
	response.StatusDescription = comment
	response.Data = data
	responseByte, _ := json.Marshal(response)
	return responseByte, nil
}


