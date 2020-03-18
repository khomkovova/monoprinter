package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"
	"time"
)

type UploadFile struct {
	Info    FileInfo
	File    multipart.File
	FilePdf *os.File
}
type FileExif struct {
	FileSize   string `json:"FileSize"`
	PageCount  int    `json:"PageCount"`
	CreateDate string `json:"CreateDate"`
}
var PDF_CONVERTER_TIMEOUT time.Duration = 30

func (uploadFile *UploadFile) getPDF() error {

	nameFile := RandStringRunes()
	convertFile, err := os.Create(nameFile)
	if err != nil {
		return errors.New("Not create Pdf file")
	}
	_, err = io.Copy(convertFile, uploadFile.File)
	if err != nil {
		return errors.New("Not copy context to Pdf file")
	}
	_ = convertFile.Close()

	arg := " --headless --convert-to pdf " + nameFile
	cmd := exec.Command("sh", "-c", "soffice "+arg) // Convert file to pdf
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Start()
	if err != nil {
		return errors.New("Not run soffice")
	}
	fmt.Println(out.String())
	// Wait for the process to finish or kill it after a timeout:
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After( PDF_CONVERTER_TIMEOUT * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("soffice failed to kill process: ", err)
			return err
		}
		log.Println("soffice process killed as timeout reached")
		return err
	case err := <-done:
		if err != nil {
			log.Fatalf("soffice process finished with error = %v", err)
			return err
		}
		//fmt.Println("process finished successfully")
		_ = os.Remove(nameFile)
	}
	err = checkSofficeReport(out.String())
	if err != nil {
		_ = os.Remove(nameFile + ".pdf")
		return err
	}
	pdfFile, err := os.Open(nameFile + ".pdf")
	if err != nil {
		return errors.New("Not open pdf file")
	}
	arg = "-json " + nameFile + ".pdf"

	cmdExifTool := exec.Command("sh", "-c", "exiftool "+arg) // Run exiftool for get file info
	var outExifTool bytes.Buffer
	cmdExifTool.Stdout = &outExifTool
	err = cmdExifTool.Run()
	if err != nil {
		return errors.New("Not run exiftool")
	}
	_ = os.Remove(nameFile + ".pdf")
	var fileExif []FileExif
	report := outExifTool.String()
	if strings.Contains(report, "File not found") {
		return errors.New("Exiftool return File not found")
	}
	err = json.Unmarshal([]byte(report), &fileExif)
	if err != nil {
		return errors.New("Not parse exiftool report")
	}
	uploadFile.Info.NumberPage = fileExif[0].PageCount
	uploadFile.Info.Size = fileExif[0].FileSize
	uploadFile.Info.UploadDate = fileExif[0].CreateDate
	uploadFile.FilePdf = pdfFile
	uploadFile.Info.Status = "STATUS_WAITING_DOWNLOAD"
	return nil

}

func RandStringRunes() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func checkSofficeReport(report string) error {
	if strings.Contains(report, "Error") {
		return errors.New("Soffice return error")
	}

	if strings.Contains(report, "Overwriting") {
		return errors.New("Soffice return Overwriting error")
	}
	return nil
}
