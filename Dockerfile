FROM golang:1.20 AS build

WORKDIR $GOPATH/src/mypackage/videoapi
COPY . .

ENV CGO_ENABLED "0"
RUN go mod tidy

WORKDIR $GOPATH/src/mypackage/videoapi/cmd/videoapi
RUN go build -o /go/bin/videoapi

FROM scratch

COPY  --from=build /go/bin/videoapi /go/bin/videoapi
ENTRYPOINT ["/go/bin/videoapi"]
