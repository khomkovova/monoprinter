package main

import (
	"MonoPrinter/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"net/http"
	"os"
)

var cache redis.Conn
var mysqlDb sql.DB
var mongoUsersCollection mgo.Collection
var mongoGridFS mgo.GridFS
var mongoPrinterCollection mgo.Collection

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
	http.HandleFunc("/api/busytime", ApiBusyTime)
	http.HandleFunc("/api/terminal/files", ApiTerminalFiles)
	l, err := net.Listen("tcp4", ":9999")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(l, nil))
}

func initMongoDb(conf config.MongodbConf) {
	session, err := mgo.Dial(conf.Host)
	if err != nil {
		fmt.Println("Don't connect to mongodb")
	}
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(conf.DatabaseName).C("test2")
	cP := session.DB(conf.DatabaseName).C("printers")
	grfs := session.DB(conf.DatabaseName).GridFS("fs")
	mongoGridFS = *grfs
	mongoUsersCollection = *c
	mongoPrinterCollection = *cP
}

func initMysql(conf config.MysqlConf) {
	var conn, err = sql.Open("mysql", conf.Username+":"+conf.Password+"@/"+conf.DatabaseName+"?charset=utf8")
	if err != nil {
		fmt.Println("Don't connect to mysql")
		return
	}
	mysqlDb = *conn
}

func initCache(conf config.RedisConf) {
	conn, err := redis.DialURL("redis://" + conf.Host)
	if err != nil {
		fmt.Println("Don't connect to redis")
		return
	}
	cache = conn
}

func initAll() error {

	go CheckOrders()
	var conf config.Configuration
	err := conf.ParseConfig()
	if err != nil {
		return err
	}
	initCache(conf.Databases.Redis)
	initMysql(conf.Databases.Mysql)
	initMongoDb(conf.Databases.MongoDb)
	return nil

}

func test() {

}
