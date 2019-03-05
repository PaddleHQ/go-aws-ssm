## Base image that contains everything needed for local dev
FROM golang:1.12.0-alpine AS development

# Install everything for development
RUN apk add bash git gcc g++

RUN go get -u golang.org/x/lint/golint

WORKDIR /app

## Copy go.mod and go.sum to install go modules and force it to use them
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download

## Copy the rest of our code in
COPY . .
