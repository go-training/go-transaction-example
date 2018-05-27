package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var globalDB *mgo.Database
var account = "appleboy"
var in chan Data

// Data struct
type Data struct {
	Account string
	Result  *chan float64
}

type currency struct {
	ID      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Amount  float64       `bson:"amount"`
	Account string        `bson:"account"`
	Code    string        `bson:"code"`
}

func pay(w http.ResponseWriter, r *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		result := make(chan float64)
		in <- Data{
			Account: account,
			Result:  &result,
		}
		for {
			select {
			case result := <-result:
				log.Printf("account: %v, result: %+v\n", account, result)
				wg.Done()
			}
		}
	}(&wg)

	wg.Wait()

	io.WriteString(w, "ok")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	in = make(chan Data)

	session, err := mgo.Dial("localhost:27017")

	if err != nil {
		panic("can't connect mongodb server")
	}

	globalDB = session.DB("logs")

	globalDB.C("bank").DropCollection()

	user := currency{Account: account, Amount: 1000.00, Code: "USD"}
	err = globalDB.C("bank").Insert(&user)

	if err != nil {
		panic("insert error")
	}

	go func(in *chan Data) {
		for {
			select {
			case data := <-*in:
				entry := currency{}
				// step 1: get current amount
				err := globalDB.C("bank").Find(bson.M{"account": data.Account}).One(&entry)

				if err != nil {
					panic(err)
				}

				//step 3: subtract current balance and update back to database
				entry.Amount = entry.Amount + 50.000
				err = globalDB.C("bank").UpdateId(entry.ID, &entry)

				if err != nil {
					panic("update error")
				}

				*data.Result <- entry.Amount
			}
		}

	}(&in)

	log.Println("Listen server on " + port + " port")
	http.HandleFunc("/", pay)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
