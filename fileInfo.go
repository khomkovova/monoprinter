package main

import (
	"errors"
	"strings"
)

type FileInfo struct {
	UniqueId string
	Filename string
	PrintingDate string

	UploadDate string
	NumberPage int
	Size string

	IdPrinter int
	Status string
}
var CountMonoPrinter = 5
func (fileInfo *FileInfo) checkInfo() error {
	if fileInfo.UniqueId == ""{
		return errors.New("Don't set UniqueId")
	}
	strings.Replace(fileInfo.Filename, "/", "", -1)
	strings.Replace(fileInfo.Filename, "..", "", -1)
	if fileInfo.Filename == ""{
		return errors.New("Don't set Filename")
	}

	var printer Printer
	printer.PrinterInfo.PrinterID = fileInfo.IdPrinter
	err := printer.getNewInfo()
	if err != nil{
		return err
	}
	_, err = printer.getPrintingTimeIndex(fileInfo.PrintingDate, fileInfo.NumberPage)
	if err != nil{
		return err
	}

	if fileInfo.UploadDate == ""{
		return errors.New("Don't set UploadDate")
	}

	if fileInfo.Status == ""{
		return errors.New("Don't set Status")
	}

	if fileInfo.NumberPage == 0{
		return errors.New("Don't set NumberPage")
	}
	if fileInfo.IdPrinter < 0 || fileInfo.IdPrinter > CountMonoPrinter{
		return errors.New("Don't set IdPrinter")
	}

	return nil
}
