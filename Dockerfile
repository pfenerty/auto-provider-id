FROM golang:1.20 as build

# WORKDIR /go/src/app

# COPY app/go.mod app/go.sum ./

# RUN go mod download

# COPY app/*.go ./

# RUN CGO_ENABLED=0 go build -o /go/bin/app

# FROM gcr.io/distroless/static-debian11:nonroot
# COPY --from=build /go/bin/app /
# CMD ["/app"]
