package main

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
)


type Printer struct {
	PrinterInfo PrinterInfo
}
type TimeLine struct {
	Date  string `bson:"Date"`
	Delay int `bson:"Delay"`
}
type PrinterInfo struct {
	PrinterID int `bson:"PrinterID"`
	TimeLine[] TimeLine `bson:"TimeLine"`
}
var PrintigDelay = time.Second * 29
var layout = "2006-01-02T15:04:05"
func (printer *Printer) addPrintingTime(printingTimeStr string, delay int) error {

	err := printer.getNewInfo()
	if err != nil {
		return err
	}

	index, err := printer.getPrintingTimeIndex(printingTimeStr, delay)
	if err != nil {
		return err
	}
	var timeLine TimeLine
	timeLine.Date = printingTimeStr
	timeLine.Delay = delay
	TL := append(printer.PrinterInfo.TimeLine[:index], append([]TimeLine{timeLine}, printer.PrinterInfo.TimeLine[index:]...)...)
	printer.PrinterInfo.TimeLine = nil
	for i := 0; i < len(TL); i++{
		printTime, _ := time.Parse(layout, TL[i].Date)
		if printTime.Before(time.Now()){
			printer.PrinterInfo.TimeLine = append(printer.PrinterInfo.TimeLine, TL[i])
		}
	}
	err = printer.setNewInfo()
	if err != nil {
		return err
	}
	return nil

}

func (printer *Printer) getNewInfo() error {
	err := mongoPrinterCollection.Find(bson.M{"PrinterID": printer.PrinterInfo.PrinterID}).One(&printer.PrinterInfo)
	if err != nil{
		err = mongoPrinterCollection.Insert(printer.PrinterInfo)
		if err != nil{
			err = errors.New("PrinterID not fount and not create")
			return err
		}
	}
	return nil
}

func (printer *Printer) setNewInfo() error {

	_, err := mongoPrinterCollection.UpdateAll(bson.M{"PrinterID": printer.PrinterInfo.PrinterID}, bson.M{"$set": printer.PrinterInfo})
	if err != nil{
		err = errors.New("Printer info not updated")
		return err
	}
	return nil
}

// get index for printing time in TimeLine and Check info delay = time for printing
func (printer *Printer)  getPrintingTimeIndex(printingTimeStr string, delay int) (int, error) {

	fmt.Println(printingTimeStr)
	printingTime, err := time.Parse(layout, printingTimeStr)
	if err != nil {
		return 0,errors.New("Don't parse printing date")
	}
	timeNow := time.Now()
	minPrintingTime := timeNow.Add(time.Minute*5)
	maxPrintingTime := timeNow.Add(time.Hour*24*7)

	if minPrintingTime.After(printingTime){
		return 0, errors.New("Can't set printingdate before 5 minut")
	}

	if maxPrintingTime.Before(printingTime){
		return 0, errors.New("Can't set printingdate after 7 day")
	}

	var TimeLine []time.Time
	for i :=0; i < len(printer.PrinterInfo.TimeLine); i++ {
		t, err := time.Parse(layout, printer.PrinterInfo.TimeLine[i].Date)
		if err != nil {
			continue
		}
		TimeLine = append(TimeLine, t)
	}
	if len(TimeLine) < 1{
		return 0, nil
	}
	if  printingTime.Add(PrintigDelay).Before(TimeLine[0]){
		return 0, nil
	}
	for i := 0; i < len(TimeLine) - 1; i++{
		d := printer.PrinterInfo.TimeLine[i].Delay
		if  (TimeLine[i].Add(PrintigDelay * time.Duration(d)).Before(printingTime) && printingTime.Add(PrintigDelay * time.Duration(delay)).Before(TimeLine[i + 1])){
			return i + 1, nil
		}
	}
	if  TimeLine[len(TimeLine)-1].Add(PrintigDelay).Before(printingTime){
		return len(TimeLine), nil
	}

	return 0, errors.New("Not added printing time to TimeLine")
}