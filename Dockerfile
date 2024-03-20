FROM golang:1.21 as build

WORKDIR /go/src/app

COPY app/go.mod app/go.sum ./

RUN go mod download

COPY app/*.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=build /go/bin/app /app
# COPY app/auto-provider-id /auto-provider-id
CMD ["/app"]
