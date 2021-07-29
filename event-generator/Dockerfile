FROM debian:buster-slim

LABEL org.opencontainers.image.source=https://github.com/estuary/demos-segmentation

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
