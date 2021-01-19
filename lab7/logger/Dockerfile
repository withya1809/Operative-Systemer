#importerer golang image
FROM golang 

#adder alle filene til gitt path
ADD . /go/src/logger

#setter current working directory
WORKDIR /go/src/logger

#go install flytter filene til /bin
RUN go install logger

ENTRYPOINT ["/go/bin/logger"]
