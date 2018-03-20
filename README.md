# go-transaction-example

Included examples to guide how to make transaction using Golang.

## Install dependency lib

```
$ go get github.com/globalsign/mgo/bson
$ go get github.com/globalsign/mgo
```

## Scenario

It demonstrate a simple server that serve pay user monery from the bank. Example steps:

1. Init bank account with an amount is 1000USD.
2. If there is a request is called to server, user will get 50$.
3. If pay is ok, calculate remain balance then update to DB.

So based on this example, The maximum times user can get pay is 20 times ( 20 X 50$ = 1000$ ), If user can be pay over 20 times, our system get fraud.

## Testing

Install [vegeta](https://github.com/tsenart/vegeta) HTTP load testing tool and library.

```
$ go get -u github.com/tsenart/vegeta
```

Sytax:

```
echo "GET http://localhost:8000" | vegeta attack -rate=1000 -duration=1s | tee results.bin | vegeta report
```

## Answer

First solution is using `sync.Mutex` to Lock other thread access this section of code until it's unlocked. You can see the [example code](./safe/safe.go).

## Alternative method

### Queue [single_queue.go](./queue/single_queue.go)

Each payment is processed one by one.Take a look at [single_queue.go](./queue/single_queue.go). I have implemented 2 channels. The first one is input channel (user acccount) and other one is output channel (Get user amount).

### Multiple queues [multiple_queue.go](./multiple_queue/multiple_queue.go)

I have 100 users and there are 10( Q ) queues are listening are numbered from 0 -> 9 ( 10 -1 ). If user X ( 0-> 99 ) I calculate what queue it should be used. My rule is simple by get modulo of X by Q.

* X = 41, Q = 10 -> The queue should be process for this request is 41 % 10 = 1 ( first queue )
* X = 33, Q = 10 -> The queue should be process for this request is 33 % 10 = 3 ( third queue )

### Optimistic concurrency control [optimistic.go](./optimistic/optimistic.go)

**If you run multiple application, please following the solutio.**

[Optimistic concurrency control](http://en.wikipedia.org/wiki/Optimistic_concurrency_control) (or `optimistic locking`) is usually implemented as an application-side method for handling concurrency, often by object relational mapping tools like Hibernate.

In this scheme, all tables have a version column or last-updated timestamp, and all updates have an extra WHERE clause entry that checks to make sure the version column hasn’t changed since the row was read. The application checks to see if any rows were affected by the UPDATE and if none were affected, treats it as an error and aborts the transaction.

## Benchmark log

Testing in [Digital Ocean](https://www.digitalocean.com/)

* OS: Ubuntu 16.04.4 x64
* Memory: 4 GB

**500 requests per second**

[safe.go](./safe/safe.go) using `sync.Mutex`

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=500 -duration=1s | tee results.bin | vegeta report
Requests      [total, rate]            500, 501.00
Duration      [total, attack, wait]    27.248468672s, 997.999728ms, 26.250468944s
Latencies     [mean, 50, 95, 99, max]  13.171447347s, 13.256585994s, 25.01617146s, 26.093162165s, 26.250468944s
Bytes In      [total, mean]            1000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:500
Error Set:
```

**500 requests per second**

[optimistic.go](./optimistic/optimistic.go) using [Optimistic concurrency control](http://en.wikipedia.org/wiki/Optimistic_concurrency_control)

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=500 -duration=1s | tee results.bin | vegeta report
Requests      [total, rate]            500, 501.00
Duration      [total, attack, wait]    5.285286131s, 997.999795ms, 4.287286336s
Latencies     [mean, 50, 95, 99, max]  1.903748023s, 1.983848904s, 4.049558826s, 4.516593338s, 5.016707396s
Bytes In      [total, mean]            1000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:500
Error Set:
```

**500 requests per second, run 60 seconds: total 30000 request**

[single_queue.go](./queue/single_queue.go) using `goroutine` + `channel`

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=500 -duration=60s | tee results.bin | vegeta report
Requests      [total, rate]            30000, 500.02
Duration      [total, attack, wait]    59.999000882s, 59.997999731s, 1.001151ms
Latencies     [mean, 50, 95, 99, max]  763.662µs, 678.816µs, 862.271µs, 1.570812ms, 66.078117ms
Bytes In      [total, mean]            60000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:30000
Error Set:
```

**500 requests per second, run 60 seconds: total 30000 request**

[multiple_queue.go](./multiple_queue/multiple_queue.go) using `goroutine` + `channel`

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=500 -duration=60s | tee results.bin | vegeta report
Requests      [total, rate]            30000, 500.02
Duration      [total, attack, wait]    59.9988601s, 59.997999803s, 860.297µs
Latencies     [mean, 50, 95, 99, max]  789.131µs, 723.723µs, 950.715µs, 1.516693ms, 49.270982ms
Bytes In      [total, mean]            60000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:30000
Error Set:
```

**500 requests per second, run 60 seconds: total 30000 request**

[optimistic_queue.go](./optimistic_queue/single_queue.go) using `goroutine` + `channel` + `Optimistic concurrency control`。Run two application in PORT `8081` and `8082`

```
$ echo "GET http://localhost:8081" | vegeta attack -rate=500 -duration=60s | tee results.bin | vegeta report
Requests      [total, rate]            30000, 500.02
Duration      [total, attack, wait]    59.999107154s, 59.997999809s, 1.107345ms
Latencies     [mean, 50, 95, 99, max]  1.297197ms, 826.466µs, 1.568221ms, 3.047957ms, 139.045488ms
Bytes In      [total, mean]            60000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:30000
Error Set:
```

**500 requests per second, run 60 seconds: total 30000 request**

[optimistic_multiple_queue.go](./optimistic_multiple_queue/multiple_queue.go) using `goroutine` + `channel` + `Optimistic concurrency control`。Run two application in PORT `8081` and `8082`

```
$ echo "GET http://localhost:8081" | vegeta attack -rate=500 -duration=60s | tee results.bin | vegeta report
Requests      [total, rate]            30000, 500.02
Duration      [total, attack, wait]    59.99868945s, 59.997999842s, 689.608µs
Latencies     [mean, 50, 95, 99, max]  924.951µs, 821.617µs, 1.2388ms, 1.978441ms, 51.268963ms
Bytes In      [total, mean]            60000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:30000
Error Set:
```

Conclustion:

|                           | max Latencies | mean Latencies | user account |
|---------------------------|---------------|----------------|--------------|
| sync lock                 | 26.250468944s | 13.171447347s  | 1            |
| optimistic lock           | 5.016707396s  | 1.903748023s   | 1            |
| single queue              | 66.078117ms   | 763.662µs      | 1            |
| multiple queue            | 49.270982ms   | 789.131µs      | 100          |
| optimistic single queue   | 139.045488ms  | 1.297197ms     | 1            |
| optimistic multiple queue | 51.268963ms   | 924.951µs      | 100          |


ref: [PostgreSQL anti-patterns: read-modify-write cycles](https://blog.2ndquadrant.com/postgresql-anti-patterns-read-modify-write-cycles/)
