package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Databases DatabasesConf `json:"Databases"`
	GCP GCPConf `json:"GCP"`

}

type GCPConf struct {
	BucketUsersFiles string 	`json:"BucketUsersFiles"`
}

type DatabasesConf struct {
	Mysql MysqlConf 	`json:"Mysql"`
	MongoDb MongodbConf `json:"MongoDb"`
	Redis RedisConf 	`json:"Redis"`
}

type MysqlConf struct {
	DatabaseName string `json:"DatabaseName"`
	Host string			`json:"Host"`
	Username string 	`json:"Username"`
	Password string		`json:"Password"`

}
type MongodbConf struct {
	DatabaseName string `json:"DatabaseName"`
	Host string			`json:"Host"`
	Username string 	`json:"Username"`
	Password string		`json:"Password"`

}
type RedisConf struct {
	DatabaseName string `json:"DatabaseName"`
	Host string			`json:"Host"`
	Username string 	`json:"Username"`
	Password string		`json:"Password"`

}

func (config *Configuration) ParseConfig() error {
	data, err := ioutil.ReadFile("config/config.json")
	if err != nil{
		return err
	}
	err = json.Unmarshal([]byte(data), config)
	if err != nil{
		return err
	}
	return nil
}
