
#!/bin/bash
# service mongodb start

cp MonoPrinterConfig/liqpay_config.json liqpay/config.json  ||  { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git"  ; exit; }
cp MonoPrinterConfig/main_config.json config/config.json || { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git" && exit ; }
ls backup  || { echo -e "\e[31mFirst download backup files. Run backup_download.sh.  From s3://monoprinter/" ; exit; }

service mysql start
redis-server &
service mongodb start
# service --status-all

mysql_password=$(cat config/config.json |  python -c 'import json,sys;obj=json.load(sys.stdin);print obj["Databases"]["Mysql"]["Password"]') || { echo -e "\e[31mBad config file" ; exit ;}

mysql --user=root --password=root -e "UPDATE mysql.user set authentication_string=password('$mysql_password') where user='root'; FLUSH PRIVILEGES;" || { echo -e "\e[31mPassword to mysql didn't change"  ; exit ; }

mongorestore --db monoprinter backup/mongodb/monoprinter
mysql -u root --password=$mysql_password  < backup/mysql/monoprinter.sql
echo -e "\e[33mWaiting download dependency"
go get -u github.com/go-delve/delve/cmd/dlv
chmod 777 /go/bin/dlv
go get -d ./...
go build
./MonoPrinter &
echo -e "\e[32mOk app is build!\e[39m!"
