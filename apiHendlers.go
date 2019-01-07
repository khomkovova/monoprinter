package main

import (
	"MonoPrinter/liqpay"
	_ "database/sql"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"net/http"
)

type CredentialsSignin struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type CredentialsRegistration struct {
	Password string `json:"password"`
	Username string `json:"username"`
	Email string `json:"email"`
	NumberPhone string `json:"numberphone"`
}


func ApiMain(w http.ResponseWriter, r *http.Request){
	username := getUsernameFromCookie(r)
	if username == ""{
		fmt.Println("Bad cookie")
	}
	fmt.Println("Yor username =", username)
}

func ApiSignin(w http.ResponseWriter, r *http.Request)  {
	var creds CredentialsSignin

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var username string
	err = mysqlDb.QueryRow("SELECT username FROM users WHERE username=? AND password=?", creds.Username, creds.Password).Scan(&username)
	if err != nil || username == ""{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cookie := makeCookie(creds.Username)
	if cookie.Value == ""{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
}

func ApiSignup(w http.ResponseWriter, r *http.Request)  {

	var creds CredentialsRegistration
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var newUsers UserInfo
	newUsers.Username = creds.Username
	newUsers.Email = creds.Email
	newUsers.NumberPhone = creds.NumberPhone
	err = newUsers.checkUser()
	if err != nil{
		fmt.Println(err)
		_, _ = w.Write([]byte("This users is registered"))
		return
	}
	err = newUsers.createNewUser()
	if err != nil{
		_, _ = w.Write([]byte("Bad creds"))
		return
	}

	id, err := mysqlDb.Query("INSERT INTO users (username,password)" + "VALUES ('" + creds.Username + "','" + creds.Password + "')" )
	if err != nil{
		_, _ = w.Write([]byte("This username using or your credentials is not correct"))
		return
	}
	_ = id.Close()


	_, _ = w.Write([]byte("You success registered"))
	w.WriteHeader(http.StatusOK)

	}


func ApiGetShortUserInfo(w http.ResponseWriter, r *http.Request)  {
	fmt.Println("99999")
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}
	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil{
		fmt.Println(err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	infoJson, err := user.makeStringJsonInfo()
	if err != nil{
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte(infoJson))
	return
	}

func ApiUpdateUserInfo(w http.ResponseWriter, r *http.Request)  {
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}
	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil{
		fmt.Println(err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	user.NumberPhone = "1111111111"

	err = user.updateInfo()
	if err != nil{
		fmt.Println(err)
		return
	}
	_, _ = w.Write([]byte("info success update"))
	return

}

func ApiUploadFile(w http.ResponseWriter, r *http.Request)  {
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}


	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil{
		fmt.Println(err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if (len(r.MultipartForm.Value["json"]) == 0){
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var uploadFile UploadFile
	jsonStr := r.MultipartForm.Value["json"][0]
	err = json.Unmarshal([]byte(jsonStr), &uploadFile.Info)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("uploadfile")
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uploadFile.File = file


	err = user.addFile(uploadFile)
	if err != nil{
		fmt.Println(err)
		_, _ = w.Write([]byte("not add file"))
		return
	}
	_, _ = w.Write([]byte("Success upload"))


}

func ApiDeleteFile (w http.ResponseWriter, r *http.Request){
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}
	var user UserInfo
	user.Username = username
	err := user.getInfo()
	if err != nil{
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

func ApiLiqpayData (w http.ResponseWriter, r *http.Request){
	username := getUsernameFromCookie(r)
	if username == "" {
		_, _ = w.Write([]byte("Bad cookies"))
		return
	}
	type Count struct {
		Count int `json:"countPage"`
	}
	count := Count{}
	err := json.NewDecoder(r.Body).Decode(&count)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newOrder := liqpay.CreateNewOrder()
	newOrder.SetUsername(username)
	newOrder.SetCountMoney(count.Count)
	err = newOrder.MakeId()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderId := newOrder.GetOrderId()
	var user UserInfo
	user.Username = username
	err = user.addOrder(orderId, "wait_accept")
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = newOrder.MakeRequestData()
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestData := newOrder.GetRequestData()
	_, _ = w.Write([]byte(requestData))


}

func ApiCheckOrderId(w http.ResponseWriter, r *http.Request){
	data := r.PostFormValue("data")
	sDec, err := b64.StdEncoding.DecodeString(data)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type DataJson struct {
		OrderId string `json:"order_id"`
	}
	var dataJson DataJson
	err = json.Unmarshal([]byte(sDec), &dataJson)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
	}
	exitingOrder := liqpay.SetupExitingOrder()
	exitingOrder.SetOrderId(dataJson.OrderId)
	err, orderInfo := exitingOrder.GetOrderIdInfo()
	if orderInfo.Status == "Error" {
		w.WriteHeader(http.StatusBadRequest)
	}
	err, user, count := exitingOrder.GetUsernameAndCountMoney()
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if orderInfo.Status == "wait_accept"{
		var u UserInfo
		err, status := u.getOrderStatus(orderInfo.OrderId)
		fmt.Println(status)
		if err != nil{
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if status == "wait_accept"{
			u.Username = user
			err = u.addPage(count)
			if err != nil{
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = u.changeOrderStatus(orderInfo.OrderId, "success")

		}

	}

}

func makeCookie(username string) http.Cookie {
	sessionToken, err := uuid.NewV4()
	if err != nil {
		return http.Cookie{Name:"token", Value:""}
	}
	_, err = cache.Do("SETEX", sessionToken.String(), "1000", username)
	if err != nil {
		return http.Cookie{Name:"token", Value:""}
	}
	return http.Cookie{Name:"token", Value:sessionToken.String()}
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






