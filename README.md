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

### Multiple queues [multiple_queue.go](./multiple/multiple_queue.go)

I have 100 users and there are 10( Q ) queues are listening are numbered from 0 -> 9 ( 10 -1 ). If user X ( 0-> 99 ) I calculate what queue it should be used. My rule is simple by get modulo of X by Q.

* X = 41, Q = 10 -> The queue should be process for this request is 41 % 10 = 1 ( first queue )
* X = 33, Q = 10 -> The queue should be process for this request is 33 % 10 = 3 ( third queue )

## Benchmark log

**500 requests per second**

[safe.go](./safe/safe.go) using `sync.Mutex`

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=500 -duration=1s | tee results.bin | vegeta report
Requests      [total, rate]            500, 501.00
Duration      [total, attack, wait]    28.972837965s, 997.999ms, 27.974838965s
Latencies     [mean, 50, 95, 99, max]  14.254423593s, 14.374552947s, 26.767317492s, 27.757755763s, 27.974838965s
Bytes In      [total, mean]            1000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:500
```

**1000 requests per second, run 10 seconds: total 10000 request**

[single_queue.go](./queue/single_queue.go) using `goroutine` + `channel`

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=1000 -duration=10s | tee results.bin | vegeta report
Requests      [total, rate]            10000, 1000.10
Duration      [total, attack, wait]    11.671634607s, 9.998998s, 1.672636607s
Latencies     [mean, 50, 95, 99, max]  769.856877ms, 591.605433ms, 1.609968811s, 1.653898066s, 1.676118025s
Bytes In      [total, mean]            20000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:10000
```

**1000 requests per second, run 10 seconds: total 10000 request**

[multiple_queue.go](./multiple_queue/multiple_queue.go) using `goroutine` + `channel`

```
$ echo "GET http://localhost:8000" | vegeta attack -rate=1000 -duration=10s | tee results.bin | vegeta report
Requests      [total, rate]            10000, 1000.10
Duration      [total, attack, wait]    10.02490889s, 9.998999s, 25.90989ms
Latencies     [mean, 50, 95, 99, max]  95.235263ms, 34.493445ms, 367.340027ms, 522.060345ms, 676.246746ms
Bytes In      [total, mean]            20000, 2.00
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:10000
Error Set:
```

Conclustion:

|                | max Latencies | mean Latencies |
|----------------|---------------|----------------|
| sync lock      | 27.974838965s | 14.254423593s  |
| single queue   | 1.676118025s  | 769.856877ms   |
| multiple queue | 676.246746ms  | 95.235263ms    |
