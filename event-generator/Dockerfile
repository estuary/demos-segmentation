FROM debian:buster-slim

RUN apt-get update -y \
 && apt-get install --no-install-recommends -y \
      ca-certificates \
      git \
      golang \
 && apt-get clean

WORKDIR /connector
RUN mkdir /connector/bin

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /connector/bin/event-generator .

ENTRYPOINT ["/connector/bin/event-generator"]
