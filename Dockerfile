FROM golang:latest
RUN mkdir -p /go/src/github.com/jetuuuu/youtube2audio
COPY . /go/src/github.com/jetuuuu/youtube2audio/
WORKDIR /go/src/github.com/jetuuuu/youtube2audio/
RUN go build -o main ./app

EXPOSE 8080
EXPOSE 8081

ENTRYPOINT [ "/go/src/github.com/jetuuuu/youtube2audio/run.sh" ]