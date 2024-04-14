FROM golang:1.22.2 AS go-builder

WORKDIR /appbuild/yamb
COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /appbuild/yamb/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -v -o yamb

FROM rust:1.77.2 AS rust-builder

RUN apt-get update && apt-get install -y \
  cmake \
  libflac-dev \
  libasound2-dev \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /libbuild/spotify-dl
COPY /lib/spotify-dl ./
RUN cargo build --release && ls -la /libbuild/spotify-dl/target/release

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
  ffmpeg \
  libflac-dev \
  libasound2-dev \
  && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /appbuild/yamb/cmd/yamb yamb

COPY --from=rust-builder /libbuild/spotify-dl/target/release/spotify-dl /usr/local/bin/

RUN chmod +x /usr/local/bin/spotify-dl

EXPOSE 3000 5000

ENTRYPOINT [ "/app/yamb" ]