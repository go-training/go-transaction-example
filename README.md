# go-transaction-example

Included examples to guide how to make transaction on Mongodb using Golang.

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
