FROM golang:1.12.0

WORKDIR /go/src/github.com/decentraland/world/
ENV GO111MODULE=on
ENV GOPATH=/go
ENV GOCACHE=/root/.cache/go-build
ENV GOOS=linux

COPY . .
RUN make
RUN chmod 775 entrypoint.d/entrypoint.sh
ENTRYPOINT ["entrypoint.d/entrypoint.sh"]
