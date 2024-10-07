FROM golang:1.22.6-alpine AS build-env

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . /app/

RUN CGO_ENABLED=0 go build -o /webhook main.go

FROM alpine:3.10

COPY --from=build-env /webhook /usr/local/bin/webhook
RUN chmod +x /usr/local/bin/webhook

ENTRYPOINT ["webhook"]