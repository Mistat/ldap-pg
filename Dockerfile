FROM golang:latest as builder

ARG SOURCE_PATH=/go/app
ARG CGO_ENABLED=0
ARG GOOS=linux

WORKDIR $SOURCE_PATH
COPY . .

RUN make

RUN cp $SOURCE_PATH/bin/ldap-pg .
COPY start.sh /usr/bin/

RUN chmod +x /usr/bin/start.sh

ENTRYPOINT ["start.sh"]

CMD ["root/start"]