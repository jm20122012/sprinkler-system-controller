FROM golang:1.22.4-alpine3.20 AS build

WORKDIR /build
COPY go.mod .
COPY go.sum .
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./db ./db
RUN go mod download
RUN go build -o main ./cmd

FROM golang:1.22.4-alpine3.20 AS final

WORKDIR /app
COPY --from=build /build/main /app/main
CMD ["/app/main"]
