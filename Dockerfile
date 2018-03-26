FROM golang:onbuild

WORKDIR /app

RUN go get github.com/codegangsta/gin

COPY . ./

EXPOSE 8080
ENTRYPOINT ["gin","-i","run","main.go"]
