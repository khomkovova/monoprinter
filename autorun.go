package main

import (
	"context"
	"encoding/json"
	"github.com/khomkovova/MonoPrinter/constant"
	"github.com/khomkovova/MonoPrinter/helper"
	"github.com/khomkovova/MonoPrinter/liqpay"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

func CheckOrders() {
	type Order struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}

	for true {
		time.Sleep(time.Second * 1)

		var order Order
		var orders []Order
		result, err := mongoUsersCollection.Distinct(context.TODO(), "orders", bson.D{{}})
		if err != nil {
			_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
			continue
		}
		if result == nil {
			continue
		}
		for _, i := range result {
			if i == nil {
				continue
			}
			resp, err := bson.Marshal(i)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}

			err = bson.Unmarshal(resp, &order)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			orders = append(orders, order)
		}
		if len(orders) == 0 {
			continue
		}
		l := liqpay.SetupExitingOrder()
		for _, o := range orders {
			if o.Status == "success" {
				continue
			}
			l.SetOrderId(o.Id)
			err, orderInfo := l.GetOrderIdInfo()
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			if orderInfo.Status != "success" {
				continue
			}
			err, email, count := l.GetEmailAndCountMoney()
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			var u UserInfo
			u.Email = email
			err = u.addPage(count)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			err = u.changeOrderStatus(o.Id, "success")
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}

		}

	}

}

func ReturnPages() {
	for true {
		time.Sleep(time.Second * 2)
		var files []FileInfo
		var filesReturnPages []FileInfo
		var file FileInfo
		result, err := mongoUsersCollection.Distinct(context.TODO(), "files", bson.D{{}})
		if err != nil {
			_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
			continue
		}
		if result == nil {
			jsonByte, _ := json.Marshal(files)
			_, _ = helper.GenerateInfoMsg(string(jsonByte), "")
			continue
		}
		for _, i := range result {
			resp, err := bson.Marshal(i)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}

			err = bson.Unmarshal(resp, &file)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			files = append(files, file)
		}

		for i := 0; i < len(files); i++ {
			file := files[i]
			if ((file.Status == constant.STATUS_WAITING_FOR_RETURN_PAGES) || (file.Status == constant.STATUS_ERROR_WITH_PRINTING)) {
				filesReturnPages = append(filesReturnPages, file)
				continue
			}
			nowTime := time.Now()
			layout := "2006-01-02T15:04:05"
			PrintingDate, _ := time.Parse(layout, file.PrintingDate)
			if (PrintingDate.Add(5*time.Minute).Before(nowTime) && (file.Status != constant.STATUS_SUCCESSFUL_PRINTED && file.Status != constant.STATUS_PAGES_RETURNED)) {
				filesReturnPages = append(filesReturnPages, file)
				continue
			}
		}
		for i := 0; i < len(filesReturnPages); i++ {
			file = filesReturnPages[i]
			var userInfo UserInfo
			err := mongoUsersCollection.FindOne(context.TODO(), bson.M{"files.uniqueid": file.UniqueId}).Decode(&userInfo)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			log.Print(userInfo)
			var user UserInfo
			user.Email = userInfo.Email
			err = user.getInfo()
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}
			err = user.addPage(file.NumberPage)
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
				continue
			}

			_, err = mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"files.uniqueid": file.UniqueId, "files.idprinter": file.IdPrinter}, bson.M{"$set": bson.M{"files.$.status": constant.STATUS_PAGES_RETURNED}})
			if err != nil {
				_, _ = helper.GenerateErrorMsg(err, constant.ERROR_SERVER, "")
			}
		}

	}
}
