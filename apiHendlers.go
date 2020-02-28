package main

import (
	"MonoPrinter/liqpay"
	_ "database/sql"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	//"gopkg.in/mgo.v2/bson"
	//"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CredentialsSignin struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type CredentialsRegistration struct {
	Password    string `json:"password"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	NumberPhone string `json:"numberphone"`
}

func ApiMain(w http.ResponseWriter, r *http.Request) {
	username := getUsernameFromCookie(r)
	if username == "" {
		fmt.Println("Bad cookie")
	}
	fmt.Println("Yor username =", username)
}

func ApiSignin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	var creds CredentialsSignin

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiSignin() --- Can't parse json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Can't parse json"))
		return
	}

	var username string
	err = mysqlDb.QueryRow("SELECT username FROM users WHERE username=? AND password=?", creds.Username, creds.Password).Scan(&username)
	if err != nil || username == "" {
		log.Println("Error: ", err)
		log.Println("ApiSignin() --- Username or password is incorrect")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cookie := makeCookie(creds.Username)
	if cookie.Value == "" {
		log.Println("Error: ", err)
		log.Println("ApiSignin() --- Can't generate cookies")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
}

func ApiSignup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	var creds CredentialsRegistration
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("Can't parse json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Can't parse json"))
		return
	}
	var newUsers UserInfo
	newUsers.Username = creds.Username
	newUsers.Email = creds.Email
	newUsers.NumberPhone = creds.NumberPhone
	err = newUsers.checkUser()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiSignup() --- This username is already used")
		_, _ = w.Write([]byte("This username is already used"))
		return
	}
	err = newUsers.createNewUser()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiSignup() --- Your credentials are bad, try change password and username.")
		_, _ = w.Write([]byte("Your credentials are bad, try change password and username"))
		return
	}
	//
	// FUCK HERE IS SQL INJECTION
	// FUCK HERE IS SQL INJECTION
	// FUCK HERE IS SQL INJECTION
	// FUCK HERE IS SQL INJECTION
	// FUCK HERE IS SQL INJECTION
	// FUCK HERE IS SQL INJECTION
	// FUCK HERE IS SQL INJECTION

	id, err := mysqlDb.Query("INSERT INTO users (username,password)" + "VALUES ('" + creds.Username + "','" + creds.Password + "')")
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiSignup() --- This username exist or your credentials is not correct")
		_, _ = w.Write([]byte("This username exist or your credentials is not correct"))
		return
	}
	_ = id.Close()

	_, _ = w.Write([]byte("You are successfully registered"))
	w.WriteHeader(http.StatusOK)

}

func ApiGetShortUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	username := getUsernameFromCookie(r)
	if username == "" {
		log.Println("ApiGetShortUserInfo() --- Bad cookies. Please sign in again")
		_, _ = w.Write([]byte("Bad cookies. Please sign in again"))
		return
	}
	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiGetShortUserInfo() --- Can't get information about user")
		_, _ = w.Write([]byte("Can't get information about user"))
		return
	}

	infoJson, err := user.makeStringJsonInfo()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiGetShortUserInfo() --- Can't marshall user information into json")
		_, _ = w.Write([]byte("Can't marshall user information into json"))
		return
	}
	_, _ = w.Write([]byte(infoJson))
	return
}

func ApiUpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}
	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil {
		fmt.Println(err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	user.NumberPhone = "1111111111"

	err = user.updateInfo()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, _ = w.Write([]byte("info success update"))
	return

}

func ApiUploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	username := getUsernameFromCookie(r)
	if username == "" {
		log.Println("ApiUploadFile() --- Bad cookies. Please sign in again")
		_, _ = w.Write([]byte("Bad cookies. Please sign in again"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiUploadFile() --- Can't get information about user")
		_, _ = w.Write([]byte("Can't get information about user"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiUploadFile() --- Can't parse multipart form")
		_, _ = w.Write([]byte("Can't parse multipart form"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(r.MultipartForm.Value["json"]) == 0 {
		log.Println("Error: ", err)
		log.Println("ApiUploadFile() --- Can't parse multipart form")
		_, _ = w.Write([]byte("Can't parse multipart form"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var uploadFile UploadFile
	jsonStr := r.MultipartForm.Value["json"][0]
	err = json.Unmarshal([]byte(jsonStr), &uploadFile.Info)

	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiUploadFile() --- Can't parse information about file")
		_, _ = w.Write([]byte("Can't parse information about file"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiUploadFile() --- Can't get file date from request")
		_, _ = w.Write([]byte("Can't get file date from request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadFile.File = file

	err = user.addFile(uploadFile)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiUploadFile() --- ", err)
		_, _ = w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte("The file is successfully uploaded"))

}

func ApiDeleteFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}
	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil {
		fmt.Println(err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	uniqueid := struct {
		Uniqueid string
	}{}
	err = json.NewDecoder(r.Body).Decode(&uniqueid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = user.deleteFile(uniqueid.Uniqueid)
	if err != nil {
		fmt.Println(err)
		_, _ = w.Write([]byte("not delete file"))
		return
	}
	_, _ = w.Write([]byte("Success deleted"))
}

func ApiLiqpayData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	username := getUsernameFromCookie(r)
	if username == "" {
		log.Println("ApiLiqpayData() --- Bad cookies. Please sign in again")
		_, _ = w.Write([]byte("Bad cookies. Please sign in again"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type Count struct {
		Count int `json:"countPage"`
	}
	count := Count{}
	err := json.NewDecoder(r.Body).Decode(&count)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiLiqpayData() --- Can't parse json with count pages")
		_, _ = w.Write([]byte("Can't parse json with count pages"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newOrder := liqpay.CreateNewOrder()
	newOrder.SetUsername(username)
	newOrder.SetCountMoney(count.Count)
	err = newOrder.MakeId()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiLiqpayData() --- Can't create new order")
		_, _ = w.Write([]byte("Can't create new order"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderId := newOrder.GetOrderId()
	var user UserInfo
	user.Username = username
	err = user.getInfo()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiLiqpayData() --- Can't get order id")
		_, _ = w.Write([]byte("Can't get order id"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = user.addOrder(orderId, "wait_accept")
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiLiqpayData() --- Can't add new order")
		_, _ = w.Write([]byte("Can't add new order"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = newOrder.MakeRequestData()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiLiqpayData() --- Can't make data request")
		_, _ = w.Write([]byte("Can't make data request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestData := newOrder.GetRequestData()
	_, _ = w.Write([]byte(requestData))

}

func ApiCheckOrderId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	data := r.PostFormValue("data")
	sDec, err := b64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiCheckOrderId() --- Can't decode data")
		_, _ = w.Write([]byte("Can't decode data"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type DataJson struct {
		OrderId string `json:"order_id"`
	}
	var dataJson DataJson
	err = json.Unmarshal([]byte(sDec), &dataJson)
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiCheckOrderId() --- Can't unmarshal json")
		_, _ = w.Write([]byte("Can't unmarshal json"))
		w.WriteHeader(http.StatusBadRequest)
	}
	exitingOrder := liqpay.SetupExitingOrder()
	exitingOrder.SetOrderId(dataJson.OrderId)
	err, orderInfo := exitingOrder.GetOrderIdInfo()
	if orderInfo.Status == "Error" {
		log.Println("Error: ", err)
		log.Println("ApiCheckOrderId() --- Order status is error")
		_, _ = w.Write([]byte("Order status is error"))
		w.WriteHeader(http.StatusBadRequest)
	}
	err, user, count := exitingOrder.GetUsernameAndCountMoney()
	if err != nil {
		log.Println("Error: ", err)
		log.Println("ApiCheckOrderId() --- Can't get username and count from order id")
		_, _ = w.Write([]byte("Can't get username and count from order id"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if orderInfo.Status == "wait_accept" {
		var u UserInfo
		err, status := u.getOrderStatus(orderInfo.OrderId)
		if err != nil {
			log.Println("Error: ", err)
			log.Println("ApiCheckOrderId() --- Can't get order status")
			_, _ = w.Write([]byte("Can't get order status"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if status == "wait_accept" {
			u.Username = user
			err = u.addPage(count)
			if err != nil {
				log.Println("Error: ", err)
				log.Println("ApiCheckOrderId() --- Can't add pages to user")
				_, _ = w.Write([]byte("Can't add pages to user"))
				return
			}
			_ = u.changeOrderStatus(orderInfo.OrderId, "success")

		}

	}
}

func ApiBusyTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8888")
	username := getUsernameFromCookie(r)
	if username == "" {
		log.Println("ApiLiqpayData() --- Bad cookies. Please sign in again")
		_, _ = w.Write([]byte("Bad cookies. Please sign in again"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		//keys, _ := r.URL.Query()["uniqueid"]
		//if len(keys) > 0 {
		//	uniqueid := keys[0]
		//	file, err := mongoGridFS.OpenId(bson.ObjectIdHex(uniqueid))
		//	if err != nil {
		//		_, _ = w.Write([]byte("Not found file"))
		//		return
		//	}
		//	b, err := ioutil.ReadAll(file)
		//	if err != nil {
		//		_, _ = w.Write([]byte("Not found file"))
		//		return
		//	}
		//	_, _ = w.Write(b)
		//	return
		//}
		var files []FileInfo
		err := mongoUsersCollection.Find(nil).Distinct("files", &files)
		if err != nil {
			log.Println("Error: ", err)
			log.Println("ApiBusyTime() --- Can't search in mongodb")
			_, _ = w.Write([]byte("Can't search in mongodb"))
			w.WriteHeader(http.StatusBadRequest)
			return
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
		jsonByte, err := json.Marshal(files)
		_, _ = w.Write(jsonByte)
		return
	}
	_, _ = w.Write([]byte("{\"status\":\"Bad request type\"}"))
	return
}

func makeCookie(username string) http.Cookie {
	sessionToken, err := uuid.NewV4()
	if err != nil {
		return http.Cookie{Name: "token", Value: ""}
	}
	_, err = cache.Do("SETEX", sessionToken.String(), "1000", username)
	if err != nil {
		return http.Cookie{Name: "token", Value: ""}
	}
	return http.Cookie{Name: "token", Value: sessionToken.String()}
}

func getUsernameFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return ""
		}
		return ""
	}
	sessionToken := cookie.Value
	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		return ""
	}
	if response == nil {
		return ""
	}
	username := fmt.Sprintf("%s", response)
	return username
}
