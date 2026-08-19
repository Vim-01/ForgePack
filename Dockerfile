FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Enable CGO for sqlite3
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev
RUN go build -o forgepack_bin .

FROM python:3.11-alpine

WORKDIR /app
COPY --from=builder /app/forgepack_bin /app/forgepack_bin
COPY python/rembg_script.py /app/python/rembg_script.py

RUN apk add --no-cache ffmpeg wget build-base
RUN pip install --no-cache-dir rembg pillow

CMD ["/app/forgepack_bin"]
