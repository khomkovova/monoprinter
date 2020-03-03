package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	//"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"
	//"io"
)

type UserInfo struct {
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	NumberPhone string     `json:"numberphone"`
	NumberPages int        `json:"numberpages"`
	Orders      []Order    `json:"orders"`
	Files       []FileInfo `json:"files"`
	Password    string     `json:"password"`
	UserId    string     `json:"userid"`
	Pictures    string     `json:"pictures"`
	RegistrationTime string `json:"registration_time"`
}

type Order struct {
	Id     string
	Status string
}

func (userInfo *UserInfo) createNewUser() error {
	if userInfo.Email == "" {
		err := errors.New("Not set email")
		return err
	}
	userInfo.NumberPages = 0
	_, err := mongoUsersCollection.InsertOne(context.TODO(), userInfo)
	if err != nil {
		fmt.Println("Not insert new users")
		return err
	}
	return nil
}

func (userInfo *UserInfo) checkUser() error {
	//coll.Find(
	err := mongoUsersCollection.FindOne(context.TODO(), bson.M{"email": userInfo.Email}).Decode(&userInfo)

	if err != nil {
		return nil
	}
	err = errors.New("This user is registered")
	return err
}

func (userInfo *UserInfo) getInfo() error {
	err := mongoUsersCollection.FindOne(context.TODO(), bson.M{"email": userInfo.Email}).Decode(&userInfo)

	if err != nil {
		err = errors.New("Not found users")
		return err
	}
	return nil
}

func (userInfo *UserInfo) makeStringJsonInfo() (string, error) {
	jsonStr, err := json.Marshal(userInfo)
	if err != nil {
		err = errors.New("Didn't make string json")
		return "", err
	}
	return string(jsonStr), nil
}

func (userInfo *UserInfo) updateInfo() error {
	var user UserInfo
	err := mongoUsersCollection.FindOne(context.TODO(), bson.M{"email": userInfo.Email}).Decode(&user)
	if err != nil {
		err = errors.New("This user isn't registered")
		return err
	}

	_, err = mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"email": userInfo.Email}, bson.M{"$set": userInfo})
	if err != nil {
		fmt.Println("Error update information")
		return err
	}
	//fmt.Println(changeInfo)

	return nil
}

// Need to update because bad style, wrote
func (userInfo *UserInfo) addFile(uploadFile UploadFile) error {
	var printer Printer

	file := uploadFile.File
	err := uploadFile.getPDF()
	if err != nil {
		fmt.Println(err)
		return err
	}
	uploadFile.Info.UniqueId =  uploadFile.Info.UploadDate + "___" + userInfo.Email + "___" + strconv.Itoa(uploadFile.Info.IdPrinter) + "___" + uploadFile.Info.Filename
	err = uploadFile.Info.checkInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}
	if userInfo.NumberPages < uploadFile.Info.NumberPage {
		return errors.New("User doesn't have pages")
	}
	userInfo.NumberPages -= uploadFile.Info.NumberPage
	err = gcp_upload_file(uploadFile)
	if err != nil {
		return err
	}
	_ = file.Close()
	printer.PrinterInfo.PrinterID = uploadFile.Info.IdPrinter
	err = printer.addPrintingTime(uploadFile.Info.PrintingDate, uploadFile.Info.NumberPage)
	if err != nil {
		return err
	}

	userInfo.Files = append(userInfo.Files, uploadFile.Info)
	err = userInfo.updateInfo()
	if err != nil {
		return err

	}

	return nil
}

// Don't work
func (userInfo *UserInfo) deleteFile(fileUniqueId string) error {
	err := mongoUsersCollection.FindOne(context.TODO(), bson.M{"email": userInfo.Email}).Decode(&userInfo)
	if err != nil {
		return nil
	}
	findStatus := false
	printingDate := ""
	printerId := 0
	for _, fileInfo := range userInfo.Files {
		if fileInfo.UniqueId == fileUniqueId {
			findStatus = true
			printingDate = fileInfo.PrintingDate
			printerId = fileInfo.IdPrinter
			break
		}
	}
	if findStatus == false {
		return errors.New("Not found fileUniqueId")
	}
	_, err = mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"email": userInfo.Email}, bson.M{"$pull": bson.M{"files": bson.M{"uniqueid": fileUniqueId}}})
	if err != nil {
		return errors.New("Not deleted file from user collection")
	}

	_, err = mongoPrinterCollection.UpdateOne(context.TODO(), bson.M{"PrinterID": printerId}, bson.M{"$pull": bson.M{"TimeLine": bson.M{"Date": printingDate}}})
	if err != nil {
		return errors.New("Not deleted file from printer collection")
	}

	return nil
}

func (userInfo *UserInfo) addOrder(orderId string, status string) error {
	var o Order
	o.Id = orderId
	o.Status = status
	userInfo.Orders = append(userInfo.Orders, o)
	err := userInfo.updateInfo()
	if err != nil {
		return err
	}

	return nil
}

func (userInfo *UserInfo) changeOrderStatus(orderId string, status string) error {
	_, err := mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"orders.id": orderId}, bson.M{"$set": bson.M{"orders.$.status": status}})
	if err != nil {
		return err
	}
	return nil
}

func (userInfo *UserInfo) getOrderStatus(orderId string) (error error, status string) {
	type Order struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}

	findOptions := options.Find()
	var results []*Order
	cur, err := mongoUsersCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var order Order
		err := cur.Decode(&order)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &order)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.TODO())
	for _, o := range results {
		if o.Id == orderId {
			return nil, o.Status
		}
	}
	return errors.New("Not fount orderid"), ""
}

func (userInfo *UserInfo) addPage(page int) error {
	err := userInfo.getInfo()
	if err != nil {
		return err
	}
	_, err = mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"email": userInfo.Email}, bson.M{"$set": bson.M{"numberpages": userInfo.NumberPages + page}})
	if err != nil {
		return err
	}
	return nil
}
