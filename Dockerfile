# Build the application from source
FROM golang:1.24.1-alpine3.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

#RUN #go install github.com/pressly/goose/v3/cmd/goose@latest

RUN CGO_ENABLED=0 GOOS=linux go build -o ./main.exe ./cmd/webserver/main.go

# Development
FROM build-stage AS dev-stage

WORKDIR /app

#RUN go install github.com/air-verse/air@latest

#CMD ["air -c .air.toml"]

ENTRYPOINT ["./main.exe"]


# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the app binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/main.exe /wish-mate

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["./wish-mate"]