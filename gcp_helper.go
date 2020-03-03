package main

import (
	"context"
	"io"
	"log"
	//"os"
	"time"
	"cloud.google.com/go/storage"
	"MonoPrinter/config"
)

func gcp_upload_file(newFile UploadFile)  error {
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return err
	}
	bucketName := conf.GCP.BucketUsersFiles
	objectName := newFile.Info.UniqueId

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return  err
	}


	//f, err := os.Open("main.go")
	//if err != nil {
	//	return
	//}
	//defer f.Close()




	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(wc, newFile.FilePdf); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil

}