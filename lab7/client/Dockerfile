
#importerer golang image
FROM golang 

#adder alle filene til gitt path
ADD . /go/src/client

#Kopierer filen og legger til i path
COPY loremipsum.txt /go/src/client/loremipsum.txt

#setter current working directory
WORKDIR /go/src/client

#go install flytter filene til /bin
RUN go install client

ENTRYPOINT ["/go/bin/client"]

