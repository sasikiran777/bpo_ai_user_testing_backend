FROM golang:1.26-alpine AS build

WORKDIR /app

RUN apk add --no-cache git make g++ gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

RUN git clone --depth 1 https://github.com/ggml-org/whisper.cpp.git /tmp/whisper.cpp
RUN make -C /tmp/whisper.cpp -j
RUN cp /tmp/whisper.cpp/main /bin/whisper-cli

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ffmpeg libstdc++

RUN adduser -D -H -u 10001 app

RUN mkdir -p storage/audio

COPY --from=build /app/models ./models
COPY --from=build /bin/api /usr/local/bin/api
COPY --from=build /bin/whisper-cli /usr/local/bin/whisper-cli

RUN chown -R app:app /app
RUN chmod +x /usr/local/bin/whisper-cli
RUN ln -sf /usr/local/bin/whisper-cli /usr/local/bin/whisper

USER app

EXPOSE 8080
ENTRYPOINT ["api"]
