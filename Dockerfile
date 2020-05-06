FROM golang:1-alpine as build-env

WORKDIR /go/src/github.com/abtasty/flagship-go-sdk

COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/example examples/qa/*.go

# Run the binary
FROM alpine

EXPOSE 8080

COPY --from=build-env /go/bin/example /go/bin/example
COPY --from=build-env /go/src/github.com/abtasty/flagship-go-sdk/examples/qa/assets /examples/qa/assets

CMD ["/go/bin/example"]