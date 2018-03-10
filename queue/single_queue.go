package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var globalDB *mgo.Database
var account = "appleboy"
var in chan string
var out chan Result

// Result output
type Result struct {
	Account string
	Result  float64
}

type currency struct {
	ID      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Amount  float64       `bson:"amount"`
	Account string        `bson:"account"`
	Code    string        `bson:"code"`
}

// Random get random value
func Random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min+1) + min
}

func pay(w http.ResponseWriter, r *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		in <- account
		for {
			select {
			case result := <-out:
				fmt.Printf("%+v\n", result)
				wg.Done()
				return
			}
		}
	}(&wg)

	wg.Wait()

	io.WriteString(w, "ok")
}

func main() {
	in = make(chan string)
	out = make(chan Result)

	session, _ := mgo.Dial("localhost:27017")
	globalDB = session.DB("logs")

	globalDB.C("bank").DropCollection()

	user := currency{Account: account, Amount: 1000.00, Code: "USD"}
	err := globalDB.C("bank").Insert(&user)

	if err != nil {
		panic("insert error")
	}

	go func(in *chan string) {
		for {
			select {
			case account := <-*in:
				entry := currency{}
				// step 1: get current amount
				err := globalDB.C("bank").Find(bson.M{"account": account}).One(&entry)

				if err != nil {
					panic(err)
				}

				//step 3: subtract current balance and update back to database
				entry.Amount = entry.Amount + 50.000
				err = globalDB.C("bank").UpdateId(entry.ID, &entry)

				if err != nil {
					panic("update error")
				}

				out <- Result{
					Account: account,
					Result:  entry.Amount,
				}
			}
		}

	}(&in)

	log.Println("Listen server on 8000 port")
	http.HandleFunc("/", pay)
	http.ListenAndServe(":8000", nil)
}
