package main

import (
	"MonoPrinter/rsaparser"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	_ "database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"time"
)

//type FileInfo struct {
//	UniqueId string
//	Filename string
//	PrintingDate string
//
//	UploadDate string
//	NumberPage int
//	Size string
//
//	IdPrinter int
//	Status string
//}
const STATUS_WAITING_DOWNLOAD = "STATUS_WAITING_DOWNLOAD"
const STATUS_WAITING_DELETE_FROM_TERMINAL = "STATUS_WAITING_DELETE_FROM_TERMINAL"

func ApiTerminalFiles(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		_, _ = w.Write([]byte("Bad cookie"))
	}
	sessionToken := cookie.Value
	err, terminalId := decryptTerminalCookie(sessionToken)
	if err != nil {
		_, _ = w.Write([]byte("Bad cookie"))
		return
	}
	fmt.Println("terminalId: ", terminalId)

	if r.Method == "GET" {
		keys, _ := r.URL.Query()["uniqueid"]
		if len(keys) > 0 {
			uniqueid := keys[0]
			file, err := mongoGridFS.OpenId(bson.ObjectIdHex(uniqueid))
			if err != nil {
				_, _ = w.Write([]byte("Not found file"))
				return
			}
			b, err := ioutil.ReadAll(file)
			if err != nil {
				_, _ = w.Write([]byte("Not found file"))
				return
			}
			_, _ = w.Write(b)
			return
		}
		var files []FileInfo
		err := mongoUsersCollection.Find(nil).Distinct("files", &files)
		if err != nil {
			_, _ = w.Write([]byte("Bad request"))
			return
		}
		for i := 0; i < len(files); i++ {
			file := files[i]
			if file.IdPrinter != terminalId || (file.Status != STATUS_WAITING_DOWNLOAD && file.Status != STATUS_WAITING_DELETE_FROM_TERMINAL) {
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
		jsonByte, err := json.Marshal(files)
		_, _ = w.Write(jsonByte)
		return
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
				_, _ = w.Write([]byte("Bad request"))
				return
			}
			err = mongoUsersCollection.Update(bson.M{"files.uniqueid": uniqueid, "files.idprinter": terminalId}, bson.M{"$set": bson.M{"files.$.status": st.Status}})
			if err != nil {
				_, _ = w.Write([]byte("Not found file"))
				return
			}

			_, _ = w.Write([]byte("OK"))
		}
	}
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
	fmt.Println("Terminal token = ", plainText)
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
	fmt.Println(sEnc)
	return nil, privateKey
}

func removeFromList(slice []FileInfo, i int) []FileInfo {
	return append(slice[:i], slice[i+1:]...)
}
