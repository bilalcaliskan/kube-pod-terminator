######## Start a builder stage #######
FROM golang:1.16-alpine as builder

LABEL maintainer="bilalcaliskan <bilalcaliskan@protonmail.com>"
WORKDIR /app
COPY . .
# Download all dependencies. Dependencies will be cached and stored in vendor folder if the go.mod and go.sum files
# are not changed
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

######## Start a new stage from scratch #######
FROM alpine:latest

WORKDIR /opt/
COPY --from=builder /app/main .

CMD ["./main"]