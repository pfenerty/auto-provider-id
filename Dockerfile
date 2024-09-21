FROM golang:1.23-bookworm AS build

WORKDIR /go/src/app

COPY app/go.mod app/go.sum ./

RUN go mod download

COPY app/*.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /go/bin/app /
CMD ["/app"]