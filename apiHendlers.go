package main

import (
	"context"
	_ "database/sql"
	"github.com/khomkovova/MonoPrinter/customlogger"
	"github.com/khomkovova/MonoPrinter/customresponse"
	"github.com/khomkovova/MonoPrinter/liqpay"
	"io/ioutil"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

type GoogleOAuth struct {
	Token string `json:"token"`
}

type Tokeninfo struct {
	Email            string `json:"email"`
	Sub              string `json:"sub"`
	Name             string `json:"name"`
	Picture          string `json:"picture,omitempty"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func ApiGoogleSignin(w http.ResponseWriter, r *http.Request) {
	w = AddResponseWriterHeaders(w)
	var googleOAuth GoogleOAuth
	err := json.NewDecoder(r.Body).Decode(&googleOAuth)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + googleOAuth.Token)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	var tokenInfo Tokeninfo

	err = json.NewDecoder(resp.Body).Decode(&tokenInfo)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	if tokenInfo.Error != "" {
		logger := customlogger.New(tokenInfo.Error, customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	var newUsers UserInfo
	newUsers.Username = tokenInfo.Name
	newUsers.Email = tokenInfo.Email
	newUsers.UserId = tokenInfo.Sub
	newUsers.Pictures = tokenInfo.Picture
	newUsers.NumberPhone = "googleOAuth"
	newUsers.RegistrationTime = time.Now().Format(time.RFC3339)
	err = newUsers.checkUser()
	if err == nil {
		err = newUsers.createNewUser()
		if err != nil {
			data, _ := json.Marshal(newUsers)
			logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
			logger.Print()
			customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
			_, _ = w.Write(customResp.GetByteResponse())
			return
		}
		cookie := http.Cookie{Name: "token", Value: googleOAuth.Token}
		http.SetCookie(w, &cookie)
		return
	}
	cookie := http.Cookie{Name: "token", Value: googleOAuth.Token}
	http.SetCookie(w, &cookie)
	return
}

func ApiGetShortUserInfo(w http.ResponseWriter, r *http.Request) {
	w = AddResponseWriterHeaders(w)
	err, email := getEmailFromCookie(r)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	var user UserInfo
	user.Email = email
	err = user.getInfo()
	if err != nil {
		data, _ := json.Marshal(user)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	infoJson, err := user.makeStringJsonInfo()
	if err != nil {
		data, _ := json.Marshal(user)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	customResp := customresponse.New(customresponse.OK_STATUS, "", "", infoJson)
	_, _ = w.Write(customResp.GetByteResponse())
	return
}

func ApiUploadFile(w http.ResponseWriter, r *http.Request) {
	w = AddResponseWriterHeaders(w)
	err, email := getEmailFromCookie(r)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	var user UserInfo
	user.Email = email
	err = user.getInfo()
	if err != nil {
		data, _ := json.Marshal(user)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		data, _ := ioutil.ReadAll(r.Body)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	if len(r.MultipartForm.Value["json"]) == 0 {
		data, _ := ioutil.ReadAll(r.Body)
		logger := customlogger.New("r.MultipartForm.Value['json']) == 0 ", customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	var uploadFile UploadFile
	jsonStr := r.MultipartForm.Value["json"][0]
	err = json.Unmarshal([]byte(jsonStr), &uploadFile.Info)

	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, jsonStr)
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		data, _ := ioutil.ReadAll(r.Body)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	uploadFile.File = file

	err = user.addFile(uploadFile)
	if err != nil {
		data, _ := json.Marshal(user)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_INFO, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_USER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	customResp := customresponse.New(customresponse.OK_STATUS, "", "", "File successfully uploaded")
	_, _ = w.Write(customResp.GetByteResponse())
	return
}

func ApiLiqpayData(w http.ResponseWriter, r *http.Request) {
	w = AddResponseWriterHeaders(w)
	err, email := getEmailFromCookie(r)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	type Count struct {
		Count int `json:"countPage"`
	}
	count := Count{}
	err = json.NewDecoder(r.Body).Decode(&count)
	if err != nil {
		data, _ := ioutil.ReadAll(r.Body)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_REQUEST, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_REQUEST, "Bad Request", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	newOrder := liqpay.CreateNewOrder()
	newOrder.SetEmail(email)
	newOrder.SetCountMoney(count.Count)
	err = newOrder.MakeId()
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	orderId := newOrder.GetOrderId()
	var user UserInfo
	user.Email = email
	err = user.getInfo()
	if err != nil {
		data, _ := json.Marshal(user)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	err = user.addOrder(orderId, "wait_accept")
	if err != nil {
		data, _ := json.Marshal(user)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	err = newOrder.MakeRequestData()
	if err != nil {
		data, _ := json.Marshal(newOrder)
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_ERROR, customresponse.ERROR_SERVER, string(data))
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	requestData := newOrder.GetRequestData()
	customResp := customresponse.New(customresponse.OK_STATUS, "", "", requestData)
	_, _ = w.Write(customResp.GetByteResponse())
	return
}

func ApiBusyTime(w http.ResponseWriter, r *http.Request) {
	w = AddResponseWriterHeaders(w)
	err, _ := getEmailFromCookie(r)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	if r.Method == "GET" {

		var files []FileInfo
		var file FileInfo
		result, err := mongoUsersCollection.Distinct(context.TODO(), "files", bson.D{{}})
		if err != nil {
			logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
			logger.Print()
			customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
			_, _ = w.Write(customResp.GetByteResponse())
			return
		}
		for _, i := range result {
			resp, err := bson.Marshal(i)
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}

			err = bson.Unmarshal(resp, &file)
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, string(resp))
				logger.Print()
				continue
			}
			file.UniqueId = ""
			files = append(files, file)
		}
		for i := 0; i < len(files); i++ {
			file := files[i]
			nowTime := time.Now()
			layout := "2006-01-02T15:04:05"
			PrintingDate, _ := time.Parse(layout, file.PrintingDate)
			if (nowTime.Add(time.Minute * 1)).After(PrintingDate) && len(files) > 1 {
				files = removeFromList(files, i)
				i--
			}
		}
		jsonByte, _ := json.Marshal(files)
		customResp := customresponse.New(customresponse.OK_STATUS, "", "", string(jsonByte))
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}
	customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_REQUEST, "Bad request type", "")
	_, _ = w.Write(customResp.GetByteResponse())
	return
}

func ApiTerminal(w http.ResponseWriter, r *http.Request) {
	w = AddResponseWriterHeaders(w)
	err, _ := getEmailFromCookie(r)
	if err != nil {
		logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_COOKIES, "")
		logger.Print()
		customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_COOKIES, "Bad Cookies", "")
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	if r.Method == "GET" {

		data, err := ioutil.ReadFile("terminal/config.json")
		if err != nil {
			logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
			logger.Print()
			customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
			_, _ = w.Write(customResp.GetByteResponse())
			return
		}
		type TerminalConf struct {
			ID               int    `json:"ID"`
			Name             string `json:"Name"`
			Location         string `json:"Location"`
			LocationComments string `json:"LocationComments"`
			Comments         string `json:"Comments"`
		}
		type Terminals struct {
			Terminals []TerminalConf `json:"Terminals"`
		}
		var terminals Terminals
		err = json.Unmarshal(data, &terminals)
		if err != nil {
			logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
			logger.Print()
			customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_SERVER, "", "")
			_, _ = w.Write(customResp.GetByteResponse())
			return
		}
		jsonByte, _ := json.Marshal(terminals)
		customResp := customresponse.New(customresponse.OK_STATUS, "", "", string(jsonByte))
		_, _ = w.Write(customResp.GetByteResponse())
		return
	}

	customResp := customresponse.New(customresponse.ERROR_STATUS, customresponse.ERROR_REQUEST, "Bad request type", "")
	_, _ = w.Write(customResp.GetByteResponse())
	return
}

func getEmailFromCookie(r *http.Request) (error, string) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return errors.New("Can't take cookie from request"), ""
	}
	sessionToken := cookie.Value
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + sessionToken)
	if err != nil {
		return errors.New("Can't check google token"), ""
	}
	var tokenInfo Tokeninfo

	err = json.NewDecoder(resp.Body).Decode(&tokenInfo)
	if err != nil {
		return errors.New("Can't parse google token"), ""
	}

	return nil, tokenInfo.Email
}

//TO DO NEED TO CHECK
//func ApiCheckOrderId(w http.ResponseWriter, r *http.Request) {
//	fmt.Println("------------------------------------------------------")
//	w.Header().Set("Access-Control-Allow-Credentials", "true")
//	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
//    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8888")
//	data := r.PostFormValue("data")
//	sDec, err := b64.StdEncoding.DecodeString(data)
//	if err != nil {
//		log.Println("Error: ", err)
//		log.Println("ApiCheckOrderId() --- Can't decode data")
//		_, _ = w.Write([]byte("{\"status\" : \"error\", \"status_description\" : \"Can't decode data\"}"))
//		return
//	}
//	type DataJson struct {
//		OrderId string `json:"order_id"`
//	}
//	var dataJson DataJson
//	err = json.Unmarshal([]byte(sDec), &dataJson)
//	if err != nil {
//		log.Println("Error: ", err)
//		log.Println("ApiCheckOrderId() --- Can't unmarshal json")
//		_, _ = w.Write([]byte("{\"status\" : \"error\", \"status_description\" : \"Can't unmarshal json\"}"))
//		return
//	}
//	exitingOrder := liqpay.SetupExitingOrder()
//	exitingOrder.SetOrderId(dataJson.OrderId)
//	err, orderInfo := exitingOrder.GetOrderIdInfo()
//	if orderInfo.Status == "Error" {
//		log.Println("Error: ", err)
//		log.Println("ApiCheckOrderId() --- Order status is error")
//		_, _ = w.Write([]byte("{\"status\" : \"error\", \"status_description\" : \"Order status is error\"}"))
//		return
//	}
//	err, email, count := exitingOrder.GetEmailAndCountMoney()
//	if err != nil {
//		log.Println("Error: ", err)
//		log.Println("ApiCheckOrderId() --- Can't get email and count from order id")
//		_, _ = w.Write([]byte("{\"status\" : \"error\", \"status_description\" : \"Can't get email and count from order id\"}"))
//		return
//	}
//	//fmt.Println(orderInfo)
//	if orderInfo.Status != "success" {
//		var u UserInfo
//		err, status := u.getOrderStatus(orderInfo.OrderId)
//		if err != nil {
//			log.Println("Error: ", err)
//			log.Println("ApiCheckOrderId() --- Can't get order status")
//			_, _ = w.Write([]byte("{\"status\" : \"error\", \"status_description\" : \"Can't get order status\"}"))
//			return
//		}
//		if status == "success" {
//			u.Email = email
//			err = u.addPage(count)
//			if err != nil {
//				log.Println("Error: ", err)
//				log.Println("ApiCheckOrderId() --- Can't add pages to user")
//				_, _ = w.Write([]byte("{\"status\" : \"error\", \"status_description\" : \"Can't add pages to user\"}"))
//				return
//			}
//			err = u.changeOrderStatus(orderInfo.OrderId, "success")
//		}
//	}
//}

//func ApiDeleteFile(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Access-Control-Allow-Credentials", "true")
//	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
//    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8888")
//    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8888")
//	err, email := getEmailFromCookie(r)
//	if err != nil{
//		log.Println("Error: ", err)
//		_, _ = w.Write([]byte("{\"status\" : \"error_cookie\", \"status_description\" : \"Bad cookies. Please sign in again\"}"))
//		return
//	}
//	var user UserInfo
//	user.Email = email
//	err = user.getInfo()
//	if err != nil {
//		fmt.Println(err)
//		_, _ = w.Write([]byte(err.Error()))
//		return
//	}
//	uniqueid := struct {
//		Uniqueid string
//	}{}
//	err = json.NewDecoder(r.Body).Decode(&uniqueid)
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//	err = user.deleteFile(uniqueid.Uniqueid)
//	if err != nil {
//		fmt.Println(err)
//		_, _ = w.Write([]byte("{\"status\" : \"\", \"status_description\" : \"not delete file\"}"))
//		return
//	}
//	_, _ = w.Write([]byte("{\"status\" : \"\", \"status_description\" : \"Success deleted\"}"))
//}

func AddResponseWriterHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "https://drukbox.club")
	return w
}
