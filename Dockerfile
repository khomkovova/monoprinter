

FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y wget git gcc golang  mongodb
RUN apt-get install -y redis-server
RUN apt-get install -y python-pip
RUN pip install awscli

RUN echo "mysql-server mysql-server/root_password password root" | debconf-set-selections
RUN echo "mysql-server mysql-server/root_password_again password root" | debconf-set-selections

RUN apt-get update && \
	apt-get -y install mysql-server-5.7 && \
	mkdir -p /var/lib/mysql && \
	mkdir -p /var/run/mysqld && \
	mkdir -p /var/log/mysql && \
	chown -R mysql:mysql /var/lib/mysql && \
	chown -R mysql:mysql /var/run/mysqld && \
	chown -R mysql:mysql /var/log/mysql


# UTF-8 and bind-address
RUN sed -i -e "$ a [client]\n\n[mysql]\n\n[mysqld]"  /etc/mysql/my.cnf && \
	sed -i -e "s/\(\[client\]\)/\1\ndefault-character-set = utf8/g" /etc/mysql/my.cnf && \
	sed -i -e "s/\(\[mysql\]\)/\1\ndefault-character-set = utf8/g" /etc/mysql/my.cnf && \
	sed -i -e "s/\(\[mysqld\]\)/\1\ninit_connect='SET NAMES utf8'\ncharacter-set-server = utf8\ncollation-server=utf8_unicode_ci\nbind-address = 0.0.0.0/g" /etc/mysql/my.cnf




ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"


# CMD /bin/bash -c "ls -lah"
# CMD /bin/bash

CMD  ["/bin/bash", "-c", "cd /go/src/ && git clone https://$git_username:$git_password@github.com/khomkovova/MonoPrinter.git && cd MonoPrinter && chmod 744 ./install_all.sh  && ./install_all.sh --git_username=$git_username --git_password=$git_password --aws_access_key_id=$aws_access_key_id --aws_secret_access_key=$aws_secret_access_key --aws_region=$aws_region && /bin/bash"]
# ENTRYPOINT [""]
# WORKDIR /go/src/MonoPrinter
# CMD tail -f /dev/null
# CMD ["/bin/bash", "-c", "echo 'asdas12345'"]


