package main

import (
	"github.com/khomkovova/MonoPrinter/liqpay"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
	"context"
)

func CheckOrders()  {
	time.Sleep(time.Second * 1)

	type Order struct {
		Id string `json:"id"`
		Status string `json:"status"`
	}
	var order Order
	var orders []Order
	result, err := mongoUsersCollection.Distinct(context.TODO(), "orders", bson.D{{}})
	if err != nil {
		log.Println("Error: ", err)
		log.Println("CheckOrders() --- Can't run distinct command")
		CheckOrders()
	}

	for _, i := range result {
		resp, err := bson.Marshal(i)
		if err != nil {
			log.Println("Error: ", err)
			log.Println("CheckOrders() --- Can't marshal interface")
			continue
		}

		err = bson.Unmarshal(resp, &order)
		if err != nil {
			log.Println("Error: ", err)
			log.Println("CheckOrders() --- Can't unmarshal data")
			continue
		}
		orders = append(orders, order)
	}
	if len(orders) == 0 {
		log.Println("CheckOrders() --- Orders len is 0")
		CheckOrders()
	}
	l := liqpay.SetupExitingOrder()
	for _, o := range orders {
		if o.Status == "success"{
			continue
		}
		l.SetOrderId(o.Id)
		err, orderInfo := l.GetOrderIdInfo()
		if err != nil {
			log.Println("Error: ", err)
			log.Println("CheckOrders() --- Can't get order info")
			continue
		}
		if orderInfo.Status != "success"{
			continue
		}
		err, email, count := l.GetEmailAndCountMoney()
		if err != nil {
			log.Println("Error: ", err)
			log.Println("CheckOrders() --- Can't get email and count from order id")
			continue
		}
		var u UserInfo
		u.Email = email
		err = u.addPage(count)
		if err != nil {
			log.Println("Error: ", err)
			log.Println("CheckOrders() --- Can't add pages to user")
			continue
		}
		err = u.changeOrderStatus(o.Id, "success")
		if err != nil {
			log.Println("Error: ", err)
			log.Println("CheckOrders() --- Can't change order status")
			//CheckOrders()
			continue
		}

	}
	CheckOrders()

}
