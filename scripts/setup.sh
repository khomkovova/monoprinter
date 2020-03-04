
#!/bin/bash
# service mongodb start
echo $git_password
echo $git_username

cp MonoPrinterConfig/liqpay_config.json liqpay/config.json  ||  { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git"  ; exit; }
cp MonoPrinterConfig/gcp_config.json gcp/config.json  ||  { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git"  ; exit; }
cp MonoPrinterConfig/main_config.json config/config.json || { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git" && exit ; }
cp MonoPrinterConfig/terminalPrivateKey.key config/  ||  { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git"  ; exit; }
cp MonoPrinterConfig/terminalPublicKey.key config/  ||  { echo -e "\e[31mFirst download config files. From https://github.com/khomkovova/MonoPrinterConfig.git"  ; exit; }

ls backup  || { echo -e "\e[31mFirst download backup files. Run backup_download.sh.  From s3://monoprinter/" ; exit; }

echo -e "\e[33mWaiting download dependency"
go get -u github.com/go-delve/delve/cmd/dlv
chmod 777 /go/bin/dlv
go get -d ./...
go build
./MonoPrinter &
echo -e "\e[32mOk app is build!\e[39m!"
