FROM golang:1.12-alpine as build

RUN echo -e  http://mirrors.tencentyun.com/alpine/v3.12/main/ > /etc/apk/repositories  &&  apk add bash git

# Set the Current Working Directory inside the container
WORKDIR /app/go-drawing

# We want to populate the module cache based on the go.{mod,sum} files.
ENV GOPROXY https://goproxy.cn
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .



# Build the Go app
RUN go build -o ./bin/drawing .
RUN git clone https://xiezhiqiang%40scrmtech.com:Waj33040!@e.coding.net/zqbc-scrm-new/scrm/go_conf.git ./go_conf



FROM alpine:3.12 as prod

WORKDIR /app/go-drawing


# 设置时区为上海gca +timezone
RUN echo -e  http://mirrors.tencentyun.com/alpine/v3.12/main/ > /etc/apk/repositories  && apk add  curl bash tree tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

RUN   addgroup -g 1200 -S www && adduser -u 1200 -D -S -G www www

RUN mkdir -p /home/mosh/drawing_upload/ && chmod -R 777 /home/mosh/drawing_upload/ && chown -R www /home/mosh/drawing_upload/

RUN mkdir -p ./runtime && chmod -R 777 ./runtime



COPY --from=build /app/go-drawing/font/ /app/go-drawing/font/
COPY --from=build /app/go-drawing/bin/drawing /app/go-drawing/bin/
COPY --from=build /app/go-drawing/go_conf/prod/drawing/conf/ /app/go-drawing/conf/



# This container exposes port 8080 to the outside world
EXPOSE 8007

# Run the binary program produced by `go install`
ENTRYPOINT ["./bin/drawing"]