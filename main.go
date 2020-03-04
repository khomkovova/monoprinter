package main

import (
	"MonoPrinter/config"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"gopkg.in/mgo.v2"
	"log"
	"net"
	"net/http"
	"os"
)






var mongoUsersCollection mongo.Collection
var mongoPrinterCollection mongo.Collection
var mongoCTX context.Context

func main() {
	test()
	err := initAll()
	if err != nil {
		fmt.Println("Not init preference")
		os.Exit(1)
	}
	fsJs := http.FileServer(http.Dir("public/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsJs))
	fsCss := http.FileServer(http.Dir("public/css"))
	http.Handle("/css/", http.StripPrefix("/css/", fsCss))
	http.HandleFunc("/api/signin/google", ApiGoogleSignin)
	http.HandleFunc("/api/getshortuserinfo", ApiGetShortUserInfo)
	//http.HandleFunc("/api/updateuserinfo", ApiUpdateUserInfo)
	http.HandleFunc("/api/uploadfile", ApiUploadFile)
	//http.HandleFunc("/api/deletefile", ApiDeleteFile)
	http.HandleFunc("/api/liqpaydata", ApiLiqpayData)
	//http.HandleFunc("/api/checkorderid", ApiCheckOrderId)
	http.HandleFunc("/api/busytime", ApiBusyTime)
	http.HandleFunc("/api/terminal/files", ApiTerminalFiles)
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	l, err := net.Listen("tcp4", port)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(l, nil))
}

func initMongoDb(conf config.MongodbConf) {
	//session, err := mgo.Dial(conf.Host)
	//if err != nil {
	//	fmt.Println("Don't connect to mongodb")
	//}
	//session.SetMode(mgo.Monotonic, true)
	//c := session.DB(conf.DatabaseName).C("test2")
	//cP := session.DB(conf.DatabaseName).C("printers")
	//mongoUsersCollection = *c
	//mongoPrinterCollection = *cP
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://" + conf.Username + ":" + conf.Password + "@" + conf.Host + "/" + conf.DatabaseName + "?retryWrites=true&w=majority"))
	if err != nil { log.Fatal(err) }
	collectionUsers := client.Database("printbox").Collection("users")
	collectionPrinters := client.Database("printbox").Collection("printers")
	mongoUsersCollection = *collectionUsers
	mongoPrinterCollection = *collectionPrinters
	mongoCTX = ctx

}

func initAll() error {

	go CheckOrders()
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		return err
	}
	initMongoDb(conf.Databases.MongoDb)
	return nil

}


func test()  {

}