package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"net/http"
)
var cache redis.Conn
var mysqlDb sql.DB
var mongoUsersCollection mgo.Collection
var mongoGridFS mgo.GridFS
var mongoPrinterCollection mgo.Collection
func main() {

	fsJs := http.FileServer(http.Dir("public/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fsJs))
	fsCss := http.FileServer(http.Dir("public/css"))
	http.Handle("/css/", http.StripPrefix("/css/", fsCss))
	http.HandleFunc("/", Main)
	http.HandleFunc("/signin", Signin)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/api/signin", ApiSignin)
	http.HandleFunc("/api/signup", ApiSignup)
	http.HandleFunc("/api/getshortuserinfo", ApiGetShortUserInfo)
	//http.HandleFunc("/api/updateuserinfo", ApiUpdateUserInfo)
	http.HandleFunc("/api/uploadfile", ApiUploadFile)
	http.HandleFunc("/api/deletefile", ApiDeleteFile)
	http.HandleFunc("/api/liqpaydata", ApiLiqpayData)
	http.HandleFunc("/api/checkorderid", ApiCheckOrderId)
	l, err := net.Listen("tcp4", ":9999")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(l, nil))
}

func initMongoDb() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		fmt.Println("Don't connect to mongodb")
	}
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("test2").C("test2")
	cP := session.DB("test2").C("printers")
	grfs := session.DB("test2").GridFS("fs")
	mongoGridFS = *grfs
	mongoUsersCollection = *c
	mongoPrinterCollection = *cP
}

func initMysql()  {
	var conn, err = sql.Open("mysql", "root:Remidolov12345@@/railway?charset=utf8")
	if err != nil {
		fmt.Println("Don't connect to mysql")
		return
	}
	mysqlDb = *conn
}

func initCache() {
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		fmt.Println("Don't connect to redis")
		return
	}
	cache = conn
}

func init()  {

	go CheckOrders()
	initCache()
	initMysql()
	initMongoDb()

}