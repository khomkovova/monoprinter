package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"strconv"

	//"io"
)

type UserInfo struct {
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	NumberPhone string     `json:"numberphone"`
	NumberPages int        `json:"numberpages"`
	Orders      []Order    `json:"orders"`
	Files       []FileInfo `json:"files"`
}

type Order struct {
	Id     string
	Status string
}

func (userInfo *UserInfo) createNewUser() error {
	if userInfo.Username == "" {
		err := errors.New("Not set username")
		return err
	}
	if userInfo.Email == "" {
		err := errors.New("Not set email")
		return err
	}
	if userInfo.NumberPhone == "" {
		err := errors.New("Not set phonenumber")
		return err
	}
	userInfo.NumberPages = 0
	err := mongoUsersCollection.Insert(userInfo)
	if err != nil {
		fmt.Println("Not insert new users")
		return err
	}
	return nil
}

func (userInfo *UserInfo) checkUser() error {
	err := mongoUsersCollection.Find(bson.M{"username": userInfo.Username}).One(&userInfo)

	if err != nil {
		return nil
	}
	err = errors.New("This user is registered")
	return err
}

func (userInfo *UserInfo) getInfo() error {
	err := mongoUsersCollection.Find(bson.M{"username": userInfo.Username}).One(&userInfo)

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
	err := mongoUsersCollection.Find(bson.M{"username": userInfo.Username}).One(&user)
	if err != nil {
		err = errors.New("This user isn't registered")
		return err
	}
	changeInfo, err := mongoUsersCollection.UpdateAll(bson.M{"username": userInfo.Username}, bson.M{"$set": userInfo})
	if err != nil {
		fmt.Println("Error update information")
		return err
	}
	fmt.Println(changeInfo)

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

	//mongoFile, err := mongoGridFS.Create(uploadFile.Info.Filename)
	//if err != nil {
	//	return errors.New("Not create file")
	//}
	//ObjectId := fmt.Sprintf("%s", mongoFile.Id())
	//id := ObjectId[13 : len(ObjectId)-2]
	uploadFile.Info.UniqueId =  uploadFile.Info.UploadDate + "___" + userInfo.Username + "___" + strconv.Itoa(uploadFile.Info.IdPrinter) + "___" + uploadFile.Info.Filename
	err = uploadFile.Info.checkInfo()
	if err != nil {
		fmt.Println(err)
		return err
	}
	if userInfo.NumberPages < uploadFile.Info.NumberPage {
		return errors.New("Don't have pages")
	}
	userInfo.NumberPages -= uploadFile.Info.NumberPage
	//_, err = io.Copy(mongoFile, uploadFile.FilePdf)
	err = gcp_upload_file(uploadFile)
	if err != nil {
		fmt.Println("Not upload file")
	}
	//_ = mongoFile.Close()
	_ = file.Close()
	printer.PrinterInfo.PrinterID = uploadFile.Info.IdPrinter
	err = printer.addPrintingTime(uploadFile.Info.PrintingDate, uploadFile.Info.NumberPage)
	if err != nil {
		return err
	}

	userInfo.Files = append(userInfo.Files, uploadFile.Info)
	err = userInfo.updateInfo()
	if err != nil {
		fmt.Println("Not updateinfo file")
	}

	return nil
}

// Don't work
func (userInfo *UserInfo) deleteFile(fileUniqueId string) error {
	err := mongoUsersCollection.Find(bson.M{"username": userInfo.Username}).One(&userInfo)
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
	err = mongoUsersCollection.Update(bson.M{"username": userInfo.Username}, bson.M{"$pull": bson.M{"files": bson.M{"uniqueid": fileUniqueId}}})
	if err != nil {
		return errors.New("Not deleted file from user collection")
	}

	err = mongoPrinterCollection.Update(bson.M{"PrinterID": printerId}, bson.M{"$pull": bson.M{"TimeLine": bson.M{"Date": printingDate}}})
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
	err := mongoUsersCollection.Update(bson.M{"orders.id": orderId}, bson.M{"$set": bson.M{"orders.$.status": status}})
	if err != nil {
		return err
	}
	return nil
}

func (userInfo *UserInfo) getOrderStatus(orderId string) (error error, status string) {
	type orders struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}
	var order []orders
	err := mongoUsersCollection.Find(nil).Distinct("orders", &order)
	if err != nil {
		return err, ""
	}
	for _, o := range order {
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
	err = mongoUsersCollection.Update(bson.M{"username": userInfo.Username}, bson.M{"$set": bson.M{"numberpages": userInfo.NumberPages + page}})
	if err != nil {
		return err
	}
	return nil
}
