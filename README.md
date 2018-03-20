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
$ echo "GET http://xxxx:8000" | vegeta attack -rate=500 -duration=1s | tee results.bin | vegeta report
Requests      [total, rate]            500, 501.00
Duration      [total, attack, wait]    26.55413484s, 997.998ms, 25.55613684s
Latencies     [mean, 50, 95, 99, max]  12.72966531s, 12.672499818s, 24.193280734s, 25.313179771s, 25.558340237s
Bytes In      [total, mean]            1000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:500
```

**500 requests per second**

[optimistic.go](./optimistic/optimistic.go) using [Optimistic concurrency control](http://en.wikipedia.org/wiki/Optimistic_concurrency_control)

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=500 -connections=1 -duration=1s | tee results.bin | vegeta report
Requests      [total, rate]            500, 501.00
Duration      [total, attack, wait]    18.539144778s, 997.999ms, 17.541145778s
Latencies     [mean, 50, 95, 99, max]  10.997901288s, 12.511709164s, 17.129697249s, 17.62389672s, 18.280330431s
Bytes In      [total, mean]            1000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:500
Error Set:
```

**1000 requests per second, run 60 seconds: total 60000 request**

[single_queue.go](./queue/single_queue.go) using `goroutine` + `channel`

```
$ echo "GET http://xxxx:8000" | vegeta attack -rate=1000 -duration=60s | tee results.bin | vegeta report
Requests      [total, rate]            60000, 1000.02
Duration      [total, attack, wait]    1m0.304137396s, 59.998999s, 305.138396ms
Latencies     [mean, 50, 95, 99, max]  160.43181ms, 134.410249ms, 296.27135ms, 601.547655ms, 672.252801ms
Bytes In      [total, mean]            120000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:60000
Error Set:
```

**1000 requests per second, run 60 seconds: total 60000 request**

[multiple_queue.go](./multiple_queue/multiple_queue.go) using `goroutine` + `channel`

```
$ echo "GET http://xxxx:8000" | vegeta attack -rate=1000 -duration=60s | tee results.bin | vegeta report
Requests      [total, rate]            60000, 1000.02
Duration      [total, attack, wait]    1m0.195742549s, 59.998999s, 196.743549ms
Latencies     [mean, 50, 95, 99, max]  132.990084ms, 131.41928ms, 151.442286ms, 204.313958ms, 476.134408ms
Bytes In      [total, mean]            120000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:60000
Error Set:
```

Conclustion:

|                | max Latencies | mean Latencies |
|----------------|---------------|----------------|
| sync lock      | 25.558340237s | 12.72966531s   |
| optimistic lock| 18.539144778s | 10.997901288s  |
| single queue   | 672.252801ms  | 160.43181ms    |
| multiple queue | 476.134408ms  | 132.990084ms   |
