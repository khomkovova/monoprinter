package main

import (
	"context"
	"github.com/khomkovova/MonoPrinter/constant"
	"github.com/khomkovova/MonoPrinter/helper"
	"github.com/khomkovova/MonoPrinter/liqpay"
	"go.mongodb.org/mongo-driver/bson"
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
