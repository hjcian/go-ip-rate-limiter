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
  - [Rate Limit Strategy and Implementation Choice](#rate-limit-strategy-and-implementation-choice)
    - [In-memory Object Description](#in-memory-object-description)
  - [References](#references)
  - [TODO](#todo)

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

假設：自第一次 Request 起算，每個 IP 每分鐘僅能接受 60 個 requests

### Response if **NOT** exceed the rate limit
- 在一分鐘內第 n 個 request，X-Ratelimit-Limit-Used 則顯示 n
- n <= 60
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
- 在一分鐘內第 61 個 request 則顯示 Error Status Code 429
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

|Header name                  |Description
|-----------------------------|-----------------------------|
|X-Ratelimit-Limit-Ip         |client 的 IP|
|X-Ratelimit-Limit-Per-Minute |預設的上限 (rate limit)|
|X-Ratelimit-Limit-Remaining  |此分鐘的 fixed window 內還剩下的額度|
|X-Ratelimit-Limit-Reset      |下一次重置的 [UNIX time](https://en.wikipedia.org/wiki/Unix_time)|
|X-Ratelimit-Limit-Used       |此分鐘的 fixed window 內已使用的額度|

## Rate Limit Strategy and Implementation Choice
- 採用 [Fixed window](https://cloud.google.com/solutions/rate-limiting-strategies-techniques#techniques-enforcing-rate-limits) 策略來實作 IP rate limit
- 採用 in-memory 的物件實作基本概念

### In-memory Object Description
**`IPRateLimiter`**
- 此為一個 goroutine-safed 物件，負責管理 IP 與 `RateLimiter` 的對應
- 透過 `GetLimiter(ip)` 方法提供取得 IP 的 `RateLimiter`，得在 middleware 的 routine 中使用
- goroutine-safed 的考量則是有 **同一個新IP** 並發地存取 server 時，避免在新建 `RateLimiter` 時出現 race condition

**`RateLimiter`**
- 此為一個 goroutine-safed 物件，為每一個 IP 自己的 fixed-window limits
- 負責管理該 IP 的使用額度，並透過 `Allow()` 方法來檢查是否可用
- goroutine-safed 的考量則是有 **同一個新IP** 並發地存取 server 時，避免記數出現 race condition （可能會少記到）

## References
- [GitHub’s Rate Limiting Documentation for Developers](https://developer.github.com/v3/#rate-limiting)
- [Rate Limiting HTTP Requests in Go based on IP address](https://dev.to/plutov/rate-limiting-http-requests-in-go-based-on-ip-address-542g)
- [golang/time/rate/rate.go](https://github.com/golang/time/blob/master/rate/rate.go#L55)

## TODO
- 做一些 Load testing 了解目前的 locking implementation 對吞吐量影響多少
  ```go
  limiter := ipLimiter.GetLimiter(ip)
  isAllow, statusSnapshot := limiter.Allow()
  ```
  - `GetLimiter(ip)` 有 lock 避免 goroutines 搶 ipLimiter 物件
  - `Allow()` 有 lock 避免同個 IP 的 requests 搶內部的 limiter 物件
- 目前未限制 IP 總量或使用 LRU 策略，`ipLimiter` 物件可能會儲存過多的 IP 佔用記憶體，應考慮
  1. 有支背景執行的 goroutine 檢查 `ipLimiter` 內是否已有 IP 物件已可 reset，若可則清除
  2. 使用外部資料庫系統儲存
