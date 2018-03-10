package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var globalDB *mgo.Database
var in []chan string
var out []chan Result
var maxUser = 100
var maxThread = 10

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
		number := Random(1, maxUser)
		channelNumber := number % maxThread
		account := "user" + strconv.Itoa(number)
		in[channelNumber] <- account
		select {
		case result := <-out[channelNumber]:
			fmt.Printf("%+v\n", result)
			wg.Done()
		}
	}(&wg)

	wg.Wait()

	io.WriteString(w, "ok")
}

func main() {
	in = make([]chan string, maxThread)
	out = make([]chan Result, maxThread)

	session, _ := mgo.Dial("localhost:27017")
	globalDB = session.DB("logs")

	globalDB.C("bank").DropCollection()

	for i := range in {
		in[i] = make(chan string)
		out[i] = make(chan Result)
	}

	// create 100 user
	for i := 0; i < maxUser; i++ {
		account := "user" + strconv.Itoa(i+1)
		user := currency{Account: account, Amount: 1000.00, Code: "USD"}
		if err := globalDB.C("bank").Insert(&user); err != nil {
			panic("insert error")
		}
	}

	for i := range in {
		go func(in *chan string, i int) {
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

					out[i] <- Result{
						Account: account,
						Result:  entry.Amount,
					}
				}
			}

		}(&in[i], i)
	}

	log.Println("Listen server on 8000 port")
	http.HandleFunc("/", pay)
	http.ListenAndServe(":8000", nil)
}
