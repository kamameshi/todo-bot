FROM golang:onbuild

WORKDIR /app

COPY . ./

EXPOSE 8080
