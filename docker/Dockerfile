FROM golang:1.22.0 AS builder

WORKDIR /appbuild/yamb
COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /appbuild/yamb/cmd
RUN ls -la /appbuild/yamb/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -v -o yamb

FROM alpine:latest

WORKDIR /app

# dependencies: ffmpeg for audio processing
RUN apk --no-cache add ffmpeg

COPY --from=builder /appbuild/yamb/cmd/yamb yamb

ENTRYPOINT [ "/app/yamb" ]