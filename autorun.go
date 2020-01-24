package main

import (
	"MonoPrinter/liqpay"
	"time"
)

func CheckOrders()  {
	//fmt.Println("CheckOrders()")
	time.Sleep(time.Minute * 1)

	type orders struct {
		Id string `json:"id"`
		Status string `json:"status"`
	}
	var order []orders
	err := mongoUsersCollection.Find(nil).Distinct("orders", &order)
	if err != nil {
		CheckOrders()
	}
	l := liqpay.SetupExitingOrder()
	for _, o := range order {
		if o.Status != "wait_accept"{
			continue
		}
		l.SetOrderId(o.Id)
		err, user, count := l.GetUsernameAndCountMoney()
		if err != nil {
			continue
		}
		var u UserInfo
		u.Username = user
		err = u.addPage(count)
		if err != nil {
			continue
		}
		_ = u.changeOrderStatus(o.Id, "success")

	}
	CheckOrders()

}
