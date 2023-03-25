FROM golang:1.20 AS build

WORKDIR $GOPATH/src/mypackage/videoapi

# Only download updates if modules files have changed
ENV CGO_ENABLED "0"
COPY go.mod .
COPY go.sum .
RUN go mod download -x

# Copy rest of the code and compile
COPY . .
WORKDIR $GOPATH/src/mypackage/videoapi/cmd/videoapi
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /go/bin/videoapi

FROM scratch

COPY  --from=build /go/bin/videoapi /go/bin/videoapi
ENTRYPOINT ["/go/bin/videoapi"]
