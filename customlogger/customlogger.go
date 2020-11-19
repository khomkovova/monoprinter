package customlogger

import (
	"encoding/json"
	"fmt"
	"github.com/khomkovova/MonoPrinter/constant"
	"runtime"
	"time"
)

const LOG_ERROR_STATUS  = "ERROR"
const LOG_INFO_STATUS  = "INFO"

const LOG_SEVERITY_INFO  = "INFO"
const LOG_SEVERITY_ERROR  = "ERROR"
const LOG_SEVERITY_WARNING  = "WARNING"
const LOG_SEVERITY_CRITICAL  = "CRITICAL"

type customLogger struct {
	TimeStamp      string `json:"time_stamp"`
	Severity       string `json:"severity"`
	CallerFunction string `json:"caller_function"`
	CodeLine       string `json:"code_line"`
	Log            string `json:"log"`
	LogStatus string `json:"log_status"`
	LogDetails     string `json:"log_details"`
}

func New( log string, severity string, logStatus string, logDetails string) customLogger {
	pc, _, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	var logger customLogger
	logger.TimeStamp = time.Now().Format(constant.TIME_LAYOUT)
	logger.Severity = severity
	logger.CallerFunction = details.Name()
	logger.CodeLine = string(line)
	logger.Log = log
	logger.LogStatus = logStatus
	logger.LogDetails = logDetails
	return logger
}

func (logger *customLogger) Print() {
	data, _ := json.Marshal(logger)
	fmt.Println(string(data))
}
