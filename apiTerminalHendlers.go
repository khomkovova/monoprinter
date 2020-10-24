package main

import (
	"github.com/khomkovova/MonoPrinter/helper"
	"github.com/khomkovova/MonoPrinter/models"
	"github.com/khomkovova/MonoPrinter/rsaparser"
	_ "github.com/khomkovova/MonoPrinter/models"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	_"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"github.com/khomkovova/MonoPrinter/config"
	"github.com/khomkovova/MonoPrinter/constant"
	"go.mongodb.org/mongo-driver/bson"
)

func ApiTerminalFiles(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("terminal_token")
	if err != nil {
		responseByte, _ := helper.GenerateErrorMsg(err, "error", "Can't get cookies")
		_, _ = w.Write(responseByte)
		return
	}
	sessionToken := cookie.Value
	err, terminalId := decryptTerminalCookie(sessionToken)
	if err != nil {
		responseByte, _ := helper.GenerateErrorMsg(err, "error","Can't decrypt cookies")
		_, _ = w.Write(responseByte)
		return
	}

	if r.Method == "GET" {
		keys, _ := r.URL.Query()["uniqueid"]
		if len(keys) > 0 {
			uniqueid := keys[0]
			var conf config.Configuration
			err := conf.ParseConfig()
			if err != nil {
				responseByte, _ := helper.GenerateErrorMsg(err, "error","Can't parse config")
				_, _ = w.Write(responseByte)
				return
			}
			bucketName := conf.GCP.BucketUsersFiles
			var gcpFile models.GCPFile
			gcpFile.FileUrl = "https://storage.googleapis.com/" + bucketName + "/" + uniqueid
			jsonByte, _ := json.Marshal(gcpFile)
			responseByte, _ := helper.GenerateInfoMsg(string(jsonByte), "")
			_, _ = w.Write(responseByte)
			return
		} else {

			var files []FileInfo
			var file FileInfo
			result, err := mongoUsersCollection.Distinct(context.TODO(), "files", bson.D{{}})
			if err != nil {
				responseByte, _ := helper.GenerateErrorMsg(err, "error","Can't run distinct command")
				_, _ = w.Write(responseByte)
				return
			}
			if result == nil {
				jsonByte, _ := json.Marshal(files)
				responseByte, _ := helper.GenerateInfoMsg(string(jsonByte), "")
				_, _ = w.Write(responseByte)
				return
			}
			for _, i := range result {
				resp, err := bson.Marshal(i)
				if err != nil {
					_, _ = helper.GenerateErrorMsg(err, "error","Can't marshal interface")
					continue
				}

				err = bson.Unmarshal(resp, &file)
				if err != nil {
					_, _ = helper.GenerateErrorMsg(err, "error","Can't unmarshal data")
					continue
				}
				files = append(files, file)
			}

			for i := 0; i < len(files); i++ {
				file := files[i]
				if file.IdPrinter != terminalId || (file.Status != constant.STATUS_WAITING_DOWNLOAD && file.Status != constant.STATUS_WAITING_DELETE_FROM_TERMINAL) {
					files = removeFromList(files, i)
					i--
					continue
				}
				nowTime := time.Now()
				layout := "2006-01-02T15:04:05"
				PrintingDate, _ := time.Parse(layout, file.PrintingDate)
				if (nowTime.Add(time.Minute * 1)).After(PrintingDate) && len(files) > 1 {
					files = removeFromList(files, i)
					i--
				}
			}
			jsonByte, _ := json.Marshal(files)
			responseByte, _ := helper.GenerateInfoMsg(string(jsonByte), "")
			_, _ = w.Write(responseByte)
			return
		}
	}

	if r.Method == "PUT" {
		keys, _ := r.URL.Query()["uniqueid"]
		if len(keys) > 0 {
			uniqueid := keys[0]
			type status struct {
				Status string `json:"Status"`
			}
			var st status
			err := json.NewDecoder(r.Body).Decode(&st)
			if err != nil {
				responseByte, _ := helper.GenerateErrorMsg(err, "error","Can't decode request")
				_, _ = w.Write(responseByte)
				return
			}
			_, err = mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"files.uniqueid": uniqueid, "files.idprinter": terminalId}, bson.M{"$set": bson.M{"files.$.status": st.Status}})
			if err != nil {
				responseByte, _ := helper.GenerateErrorMsg(err, "error","Can't find file")
				_, _ = w.Write(responseByte)
				return
			}
			responseByte, _ := helper.GenerateInfoMsg("", "Status changed")
			_, _ = w.Write(responseByte)
			return
		}
	}
	responseByte, _ := helper.GenerateErrorMsg(errors.New("Bad request"), "error","Bad request")
	_, _ = w.Write(responseByte)
	return
}

func decryptTerminalCookie(cookie string) (err error, terminalId int) {
	sDec, _ := base64.StdEncoding.DecodeString(cookie)
	label := []byte("")
	hash := sha256.New()
	err, privateKey := getPrivateKey()
	if err != nil {
		return err, 0
	}
	plainText, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, sDec, label)
	if err != nil {
		return errors.New("Didn't decrypt cookie"), 0
	}
	type DecryptedCookie struct {
		TerminalId int    `json:"terminalId"`
		CreateDate string `json:"createDate"`
	}
	var decryptedCookie DecryptedCookie
	err = json.Unmarshal(plainText, &decryptedCookie)
	if err != nil {
		return errors.New("Didn't decrypt cookie"), 0
	}
	nowTime := time.Now()
	layout := "2006-01-02T15:04:05"
	createCookiedate, err := time.Parse(layout, decryptedCookie.CreateDate)
	if err != nil {
		return errors.New("Didn't decrypt cookie"), 0
	}
	if nowTime.After(createCookiedate.Add(time.Minute * 10)) {
		return errors.New("Cookie is old"), 0
	}
	terminalId = decryptedCookie.TerminalId
	return nil, terminalId
}

func getPrivateKey() (err error, key *rsa.PrivateKey) {
	data, err := ioutil.ReadFile("config/terminalPrivateKey.key")
	if err == nil {
		privateKey, err := rsaparser.ParseRsaPrivateKeyFromPemStr(string(data))
		if err == nil {
			return nil, privateKey
		}
	}

	privateKey, publicKey := rsaparser.GenerateRsaKeyPair()
	strPublicKey, err := rsaparser.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		return errors.New("Doesn't create key"), nil
	}
	err = ioutil.WriteFile("config/terminalPrivateKey.key", []byte(rsaparser.ExportRsaPrivateKeyAsPemStr(privateKey)), 0644)
	_ = ioutil.WriteFile("config/terminalPublicKey.key", []byte(strPublicKey), 0644)
	message := []byte("{\"terminalId\":1, \"createDate\":\"2019-03-29T10:10:10\"}")
	label := []byte("")
	hash := sha256.New()
	ciphertext, _ := rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, label)

	sEnc := base64.StdEncoding.EncodeToString(ciphertext)
	_ = ioutil.WriteFile("config/terminalToken.key", []byte(sEnc), 0644)
	log.Println(sEnc)
	return nil, privateKey
}

func removeFromList(slice []FileInfo, i int) []FileInfo {
	return append(slice[:i], slice[i+1:]...)
}
