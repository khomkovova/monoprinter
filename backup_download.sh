aws_access_key_id=$(cat MonoPrinterConfig/main_config.json |  python -c 'import json,sys;obj=json.load(sys.stdin);print obj["AWS"]["S3"]["aws_access_key_id"]')
aws_secret_access_key=$(cat MonoPrinterConfig/main_config.json |  python -c 'import json,sys;obj=json.load(sys.stdin);print obj["AWS"]["S3"]["aws_secret_access_key"]')
aws_region=$(cat MonoPrinterConfig/main_config.json |  python -c 'import json,sys;obj=json.load(sys.stdin);print obj["AWS"]["S3"]["aws_region"]')

aws configure set aws_access_key_id $aws_access_key_id 
aws configure set aws_secret_access_key $aws_secret_access_key
aws configure set region $aws_region
aws s3 cp --recursive  s3://monoprinter/ .
