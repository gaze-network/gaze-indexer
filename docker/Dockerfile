FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

ENV GOOS=linux
ENV CGO_ENABLED=0

RUN go build \
          -o main ./main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata


COPY --from=builder /app/main .

# You can set `TZ` environment variable to change the timezone

CMD ["/app/main", "run"]
