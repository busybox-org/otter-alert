#docker build --network host --rm --build-arg APP_ROOT=/go/src/otteralter -t 172.16.127.171:10001/otteralter:<tag> -f Dockerfile .
#0 ----------------------------
FROM golang:1.19
ARG  APP_ROOT
WORKDIR ${APP_ROOT}
COPY ./ ${APP_ROOT}

ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.io/,direct"
ENV PATH=$GOPATH/bin:$PATH

# install upx
RUN sed -i "s/deb.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list \
  && sed -i "s/security.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list \
  && apt-get update \
  && apt-get install upx musl-dev git -y

# build code
RUN GO_VERSION=`go version|awk '{print $3" "$4}'` \
  && GIT_URL=`git remote -v|grep push|awk '{print $2}'` \
  && GIT_BRANCH=`git rev-parse --abbrev-ref HEAD` \
  && GIT_COMMIT=`git rev-parse HEAD` \
  && VERSION=`git describe --tags --abbrev=0` \
  && GIT_LATEST_TAG=`git describe --tags --abbrev=0` \
  && BUILD_TIME=`date +"%Y-%m-%d %H:%M:%S %Z"` \
  && go mod tidy \
  && go get \
  && CGO_ENABLED=0 GOOS=linux go build -ldflags \
  "-w -s -X 'github.com/xmapst/otteralert.Version=${VERSION}' \
  -X 'github.com/xmapst/otteralert.GoVersion=${GO_VERSION}' \
  -X 'github.com/xmapst/otteralert.GitUrl=${GIT_URL}' \
  -X 'github.com/xmapst/otteralert.GitBranch=${GIT_BRANCH}' \
  -X 'github.com/xmapst/otteralert.GitCommit=${GIT_COMMIT}' \
  -X 'github.com/xmapst/otteralert.GitLatestTag=${GIT_LATEST_TAG}' \
  -X 'github.com/xmapst/otteralert.BuildTime=${BUILD_TIME}'" \
  -o otter-alert \
  cmd/otter-alert.go \
  && strip --strip-unneeded otter-alert \
  && upx --lzma otter-alert

FROM alpine:latest
ARG APP_ROOT
WORKDIR /app/
COPY --from=0 ${APP_ROOT}/otter-alert .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add --no-cache openssh jq curl busybox-extras \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/app/otter-alert"]