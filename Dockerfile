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
  "-w -s -X 'main.Version=${VERSION}' \
  -X 'main.GoVersion=${GO_VERSION}' \
  -X 'main.GitUrl=${GIT_URL}' \
  -X 'main.GitBranch=${GIT_BRANCH}' \
  -X 'main.GitCommit=${GIT_COMMIT}' \
  -X 'main.GitLatestTag=${GIT_LATEST_TAG}' \
  -X 'main.BuildTime=${BUILD_TIME}'" \
  -o main \
  && strip --strip-unneeded main \
  && upx --lzma main

FROM alpine:latest
ARG APP_ROOT
WORKDIR /app/
COPY --from=0 ${APP_ROOT}/main .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add --no-cache openssh jq curl busybox-extras \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/app/main"]