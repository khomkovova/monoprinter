
#!/bin/bash

for i in "$@"
do
case $i in
    -gp=*|--git_password=*)
    git_password="${i#*=}"

    ;;
    -gu=*|--git_username=*)
    git_username="${i#*=}"
    ;;
    -aak=*|--aws_access_key_id=*)
    aws_access_key_id="${i#*=}"
    ;;
    -asa=*|--aws_secret_access_key=*)
    aws_secret_access_key="${i#*=}"
    ;;
    -ar=*|--aws_region=*)
    aws_access_key_id="${i#*=}"
    ;;

    
    *)
          printf 'SET ALL OPTIONS:\n git_username \n git_password \n aws_access_key_id \n aws_secret_access_key \n aws_region \n'  # unknown option
    ;;
esac
done
if [ "$git_password" == "" ] 
then 
	echo "SET git_password"
	exit 0
fi

if [ "$git_username" == "" ] 
then 
	echo "SET git_username"
	exit 0
fi

if [ "$aws_access_key_id" == "" ] 
then 
	echo "SET aws_access_key_id"
	exit 0
fi

if [ "$aws_secret_access_key" == "" ] 
then 
	echo "SET aws_secret_access_key"
	exit 0
fi

if [ "$aws_region" == "" ] 
then 
	echo "SET aws_region"
	exit 0
fi

# git_username
# git_password
# aws_access_key_id
# aws_secret_access_key
# aws_region
aws configure set aws_access_key_id $aws_access_key_id 
aws configure set aws_secret_access_key $aws_secret_access_key
aws configure set region $aws_region

service mongodb start
service mysql start
service --status-all

# cd /go/src && git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinter.git 
# cd /go/src/MonoPrinter 
git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinterConfig.git
ls -lah
cp MonoPrinterConfig/liqpay_config.json liqpay/config.json
cp MonoPrinterConfig/main_config.json config/config.json
mysql_password=$(cat config/config.json |  python -c 'import json,sys;obj=json.load(sys.stdin);print obj["Databases"]["Mysql"]["Password"]')

aws s3 cp --recursive  s3://monoprinter/ . 
ls -lah
mongorestore --db monoprinter backup/mongodb/monoprinter
mysql -u root -p   < backup/mysql/monoprinter.sql
go build
exit 0
