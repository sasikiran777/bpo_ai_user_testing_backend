FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.21

RUN adduser -D -H -u 10001 app
USER app

COPY --from=build /bin/api /usr/local/bin/api

EXPOSE 8080
ENTRYPOINT ["api"]
