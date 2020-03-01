package liqpay

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pasztorpisti/qs"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

var PublicKey string
var PrivateKey string
var OrderEncryptKey  []byte

type Configuration struct {
	PublicKey string `json:"PublicKey"`
	PrivateKey string `json:"PrivateKey"`
	OrderEncryptKey string `json:"OrderEncryptKey"`
}
func init() {
	data, err := ioutil.ReadFile("liqpay/config.json")
	if err != nil{
		fmt.Println("LiqPay config not read")
		os.Exit(2)
	}
	var config Configuration
	err = json.Unmarshal([]byte(data), &config)
	if err != nil{
		fmt.Println("LiqPay config not parse")
		os.Exit(2)
	}
	PublicKey = config.PublicKey
	PrivateKey = config.PrivateKey
	OrderEncryptKey = []byte(config.OrderEncryptKey)
	return
}
func CreateNewOrder() NewOrder {
	return &Order{}
}

func SetupExitingOrder() ExitingOrder {
	return &Order{}
}

type NewOrder interface {
	SetUsername(username string)
	SetCountMoney(countMoney int)
	MakeId() error
	GetOrderId() string
	MakeRequestData() error
	GetRequestData() string
}

type ExitingOrder interface {
	SetOrderId(orderId string)
	GetOrderIdInfo() (error, OrderInfo)
	GetUsernameAndCountMoney() (error error, username string, countMoney int)
}

type OrderInfo struct {
	Status string `json:"status"`
	Result string `json:"result"`
	OrderId string `json:"order_id"`
	Amount int `json:"amount"`
}

type Order struct {
	 id string
	 user string
	 countMoney int
	 requestData string
}

func (order *Order) SetUsername(username string)  {
	order.user = username
}

func (order *Order) SetCountMoney(countMoney int)  {
	order.countMoney = countMoney
}

func (order *Order) SetOrderId(orderId string)  {
	order.id = orderId
}

func (order *Order) GetUsername() string{
	return order.user
}

func (order *Order) GetCountMoney() int{
	return order.countMoney
}

func (order *Order) GetOrderId() string{
	return order.id
}

func (order *Order) GetRequestData() string {
	return order.requestData
}

// First need runt SetUsername() and SetCountMoney()
func (order *Order)MakeId() error {

	if order.user == "" {
		return errors.New("Username is empty")
	}
	if order.countMoney == 0 {
		return errors.New("CountMoney is empty")
	}
	t := time.Now()
	plainId := `{"user":"` + order.user + `", "count":` + strconv.Itoa(order.countMoney) + `, "time":` + strconv.Itoa(int(t.UnixNano()/1000)) + `}`
	err, cipherId := order.ecryptId(plainId)
	if err != nil{
		return err
	}
	order.id = cipherId
	return nil
}

func (order * Order)ecryptId(plainId string) (error, string) {
	c, err := aes.NewCipher(OrderEncryptKey)
	if err != nil {
		return  err, ""
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err, ""
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err, ""
	}
	sEnc := b64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plainId), nil))
	return  nil, sEnc
}

func (order * Order)decryptId(cipherId string) (error, string) {
	sDec, err := b64.StdEncoding.DecodeString(cipherId)
	if err != nil{
		return err, ""
	}
	c, err := aes.NewCipher(OrderEncryptKey)
	if err != nil {
		return err, ""
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err, ""
	}
	nonceSize := gcm.NonceSize()
	if len(sDec) < nonceSize {
		return errors.New("ciphertext too short"), ""
	}
	nonce, ciphertext := sDec[:nonceSize], sDec[nonceSize:]
	plainId, err :=  gcm.Open(nil, nonce, ciphertext, nil)
	return err, string(plainId)
}


// First need run MakeId() for create Id
func (order *Order)MakeRequestData() (error) {

	OrderData := map[string]interface{}{
		"action": "pay",
		"version": 3,
		"public_key": PublicKey,
		"amount": order.countMoney,
		"currency": "UAH",
		"description": "",
		"order_id": order.id,
	}
	b, err := json.Marshal(OrderData)
	if err != nil {
		return err
	}
	data := b64.StdEncoding.EncodeToString(b)

	h := sha1.New()
	h.Write([]byte(PrivateKey + data + PrivateKey))
	bs := h.Sum(nil)

	signature := base64.StdEncoding.EncodeToString(bs)
	order.requestData = `{"data":"` + data + `", "signature":"` + signature + `"}`

	return nil
}

// First need SetOrderId()
func (order *Order)GetUsernameAndCountMoney() (error error, username string, countMoney int){
	err, plainOrderId := order.decryptId(order.id)
	if err != nil{
		return err, "", 0
	}
	type OrderIdJson struct {
		Username string `json:"user"`
		Count int `json:"count"`
	}
	var orderIdJson OrderIdJson
	err = json.Unmarshal([]byte(plainOrderId), &orderIdJson)
	if err != nil{
		return err, "", 0
	}
	return nil, orderIdJson.Username, orderIdJson.Count
}

type formData struct {
	Data      string `json:"data"`
	Signature string `json:"signature"`
}
// First need SetOrderId()
func (order * Order)GetOrderIdInfo() (error, OrderInfo){
	var oI OrderInfo
	j := map[string]interface{}{
		"action":     "status",
		"version":    3,
		"public_key": PublicKey,
		"order_id":   order.id,
	}
	st, err := json.Marshal(j)
	if err != nil {
		return err, oI
	}
	b := base64.StdEncoding.EncodeToString([]byte(st))
	var f  formData
	f.Data = b

	h := sha1.New()
	h.Write([]byte(PrivateKey + b + PrivateKey))
	bs := h.Sum(nil)

	f.Signature = base64.StdEncoding.EncodeToString(bs)
	data, err := qs.Marshal(f)
	if err != nil {
		return err, oI
	}

	req, err := http.NewRequest("POST", "https://www.liqpay.ua/api/request", bytes.NewBufferString(data))
	if err != nil {
		return err, oI
	}
	// Do the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err, oI
	}
	byte, _ :=ioutil.ReadAll(response.Body)

	err = json.Unmarshal(byte, &oI)
	if err != nil {
		return err, oI
	}
	return nil, oI
}