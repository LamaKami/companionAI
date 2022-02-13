FROM golang:1.17-alpine

WORKDIR /companionAI

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY dockerManager ./dockerManager
COPY docs ./docs
COPY helper ./helper
COPY utils ./utils
COPY groups ./groups
COPY *.go ./

RUN go build -o /companionAI

ARG value
ENV envValue=$value

CMD ["sh", "-c", "./companionAI ${envValue}"]