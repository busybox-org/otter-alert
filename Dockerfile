#docker build --network host --rm --build-arg APP_ROOT=/go/src/otteralter -t 172.16.127.171:10001/otteralter:<tag> -f Dockerfile .
#0 ----------------------------
FROM golang:latest AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download


# Copy the go source
COPY cmd/main.go cmd/main.go
COPY internal/ internal/

# build code
RUN CGO_ENABLED=0 go build -a -trimpath -ldflags '-w -s' -o otter-alert cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /workspace/otter-alert .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add --no-cache openssh jq curl busybox-extras \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/app/otter-alert"]