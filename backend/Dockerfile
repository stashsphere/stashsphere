FROM golang:1.25-trixie as builder

RUN apt-get update
RUN apt-get install -y libmagic-dev
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /usr/local/bin/
RUN go build -v -o /usr/local/bin/backend

FROM debian:trixie
RUN apt-get update
RUN apt-get install -y libmagic-dev

RUN mkdir -p /usr/local/bin/
COPY --from=builder /usr/local/bin/backend /usr/local/bin/.

CMD ["/usr/local/bin/backend"]
