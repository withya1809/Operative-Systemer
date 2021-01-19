#importerer golang image
FROM golang

#adder alle filene til gitt path
ADD . /go/src/server

#setter current working directory
WORKDIR /go/src/server

#go install flytter filene til /bin
RUN go install server

ENTRYPOINT ["/go/bin/server"]


