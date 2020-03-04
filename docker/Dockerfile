

FROM ubuntu:18.04
RUN date
RUN apt-get update -qq
RUN apt-get upgrade -qq
RUN apt-get -y install -qq  apt-utils
RUN apt-get install -y libreoffice 
RUN apt-get install -y wget
RUN apt-get install -y git
RUN apt-get install -y gcc
#RUN add-apt-repository ppa:longsleep/golang-backports
RUN apt-get install -y golang-go
RUN apt-get install -y python-pip
RUN apt-get install -y net-tools
RUN apt-get install -y vim
RUN apt-get install -y lsof
RUN pip install awscli
RUN apt-get install -y exiftool 

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

CMD cd /go/src/ && git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinter.git && ls && cd MonoPrinter/script && chmod 744 ./setup.sh  && ./setup.sh --git_username=$git_username --git_password=$git_password  && /bin/bash

# CMD  ["/bin/bash", "-c", "cd /go/src/ && git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinter.git && cd MonoPrinter && chmod 744 ./install_all.sh  && ./install_all.sh --git_username=$git_username --git_password=$git_password --aws_access_key_id=$aws_access_key_id --aws_secret_access_key=$aws_secret_access_key --aws_region=$aws_region && /bin/bash"]
# ENTRYPOINT [""]
# WORKDIR /go/src/MonoPrinter
# CMD tail -f /dev/null
# CMD ["/bin/bash", "-c", "echo 'asdas12345'"]


