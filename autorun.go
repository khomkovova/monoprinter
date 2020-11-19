package main

import (
	"context"
	"github.com/khomkovova/MonoPrinter/constant"
	"github.com/khomkovova/MonoPrinter/customlogger"
	"github.com/khomkovova/MonoPrinter/customresponse"
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
			logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
			logger.Print()
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
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}

			err = bson.Unmarshal(resp, &order)
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
				logger.Print()
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
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}
			if orderInfo.Status != "success" {
				continue
			}
			err, email, count := l.GetEmailAndCountMoney()
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}
			var u UserInfo
			u.Email = email
			err = u.addPage(count)
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}
			err = u.changeOrderStatus(o.Id, "success")
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
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
			logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
			logger.Print()
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
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}

			err = bson.Unmarshal(resp, &file)
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
				logger.Print()
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
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_WARNING, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}
			log.Print(userInfo)
			var user UserInfo
			user.Email = userInfo.Email
			err = user.getInfo()
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}
			err = user.addPage(file.NumberPage)
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
				continue
			}

			_, err = mongoUsersCollection.UpdateOne(context.TODO(), bson.M{"files.uniqueid": file.UniqueId, "files.idprinter": file.IdPrinter}, bson.M{"$set": bson.M{"files.$.status": constant.STATUS_PAGES_RETURNED}})
			if err != nil {
				logger := customlogger.New(err.Error(), customlogger.LOG_SEVERITY_CRITICAL, customresponse.ERROR_SERVER, "")
				logger.Print()
			}
		}

	}
}
