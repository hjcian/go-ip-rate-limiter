# IP Rate limit practice
This repo is an IP rate limit middleware practice by Go.

## Table of Contents
- [IP Rate limit practice](#ip-rate-limit-practice)
  - [Table of Contents](#table-of-contents)
  - [Install](#install)
  - [Run Test](#run-test)
  - [Run Server](#run-server)
  - [Usage](#usage)
    - [Response if **NOT** exceed the rate limit](#response-if-not-exceed-the-rate-limit)
    - [Response if **EXCEED** the rate limit](#response-if-exceed-the-rate-limit)
    - [Response Header Explanation](#response-header-explanation)
  - [References](#references)

## Install
> System Prerequisites
> - Go 1.15 or latter

```shell
git clone https://github.com/hjcian/go-ip-rate-limiter.git
cd go-ip-rate-limiter
go install
```

## Run Test
```shell
go test ./... -v
```

## Run Server
```shell
go run main.go
```

## Usage

### Response if **NOT** exceed the rate limit
```shell
$ curl -I http://localhost:8080/foobar
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Ratelimit-Limit-Ip: ::1
X-Ratelimit-Limit-Per-Minute: 60
X-Ratelimit-Limit-Remaining: 59
X-Ratelimit-Limit-Reset: 1604828729
X-Ratelimit-Limit-Used: 1
Date: Sun, 08 Nov 2020 09:47:00 GMT
Content-Length: 13
```

### Response if **EXCEED** the rate limit
```shell
$ curl -I http://localhost:8080/foobar
HTTP/1.1 429 Too Many Requests
X-Ratelimit-Limit-Ip: ::1
X-Ratelimit-Limit-Per-Minute: 60
X-Ratelimit-Limit-Remaining: 0
X-Ratelimit-Limit-Reset: 1604828820
X-Ratelimit-Limit-Used: 60
Date: Sun, 08 Nov 2020 09:47:21 GMT
```

### Response Header Explanation
> 假設：自第一次 Request 起算，每個 IP 每分鐘僅能接受 60 個 requests

|Header name                  |Description
|-----------------------------|-----------------------------|
|X-Ratelimit-Limit-Ip         |client 的 IP|
|X-Ratelimit-Limit-Per-Minute |預設的上限 (rate limit)|
|X-Ratelimit-Limit-Remaining  |此分鐘這個 window 內還剩下的額度|
|X-Ratelimit-Limit-Reset      |下一次重置的 [UNIX time](https://en.wikipedia.org/wiki/Unix_time)|
|X-Ratelimit-Limit-Used       |此分鐘這個 window 內還可以使用的額度|

## References
- [GitHub’s Rate Limiting Documentation for Developers](https://developer.github.com/v3/#rate-limiting)
- [Rate Limiting HTTP Requests in Go based on IP address](https://dev.to/plutov/rate-limiting-http-requests-in-go-based-on-ip-address-542g)
- [golang/time/rate/rate.go](https://github.com/golang/time/blob/master/rate/rate.go#L55)