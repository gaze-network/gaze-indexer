FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download

COPY ./ ./

ENV GOOS=linux
ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/go/pkg/mod/ \
    go build -o main ./main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata


COPY --from=builder /app/main .

# You can set TZ indentifier to change the timezone, See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
# ENV TZ=US/Central

CMD ["/app/main", "run"]
