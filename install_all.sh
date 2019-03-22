

FROM ubuntu:18.04
FROM mysql



# ARG git_username
# ARG git_password
# ARG aws_access_key_id
# ARG aws_secret_access_key
# ARG aws_region


RUN apt-get update
RUN apt-get install -y wget git gcc golang  mongodb
RUN apt-get install -y redis-server
RUN apt-get install -y python-pip
RUN pip install awscli

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
# ---------------------------------------------------------------------
# RUN aws configure set aws_access_key_id $aws_access_key_id
# RUN aws configure set aws_secret_access_key $aws_secret_access_key
# RUN aws configure set region $aws_region



# ENV git_username
# ENV git_password
# ENV aws_access_key_id
# ENV aws_secret_access_key
# ENV aws_region

# RUN cd /go/src && git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinter.git 
# WORKDIR /go/src/MonoPrinter 
# RUN git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinterConfig.git
# RUN ls -lah
# RUN cp MonoPrinterConfig/liqpay_config.json liqpay/config.json
# RUN cp MonoPrinterConfig/main_config.json config/config.json
# RUN mysql_password=$(cat config/config.json |  python -c 'import json,sys;obj=json.load(sys.stdin);print obj["Databases"]["Mysql"]["Password"]')


# RUN aws s3 cp --recursive  s3://monoprinter/ . 

# RUN ls -lah
# RUN mongorestore --db monoprinter backup/mongodb/monoprinter
# RUN mysql -u root -p   < backup/mysql/monoprinter.sql

# RUN go build 

# ----------------------------------------------------------------------
# ENTRYPOINT ["/bin/bash", "-c", "./llll$git_username"]




CMD  ["/bin/bash"]
ENTRYPOINT ["/bin/bash", "-c", "git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinter.git && cd MonoPrinter && chmod 744 ./install_all.sh  && ./install_all.sh --git_username=$git_username --git_password=$git_password --aws_access_key_id=$aws_access_key_id --aws_secret_access_key=$aws_secret_access_key --aws_region=$aws_region"]
# WORKDIR /go/src/MonoPrinter
# CMD ["/bin/bash", "-c", " "]
# CMD ["/bin/bash", "-c", "echo 'asdas12345'"]


