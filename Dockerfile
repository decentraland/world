# BUILD STAGE
FROM golang:1.13 as builder

WORKDIR /usr/src/app
ENV GO111MODULE=on
ENV GOPATH=/go
ENV GOCACHE=/root/.cache/go-build
ENV GOOS=linux

# Install app dependencies
COPY . .

RUN apt-get update && apt-get install -y \
    libssl-dev

RUN make buildall
RUN go install -v ./...

# DEPLOY STAGE
FROM golang:1.13

RUN apt-get update && apt-get install -y \
    libssl-dev

WORKDIR /root

# Bundle app source
COPY --from=builder /usr/src/app .
