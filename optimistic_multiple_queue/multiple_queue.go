package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var globalDB *mgo.Database
var in []chan Data
var maxUser = 100
var maxThread = 10

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
	Version int           `bson:"version"`
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
		result := make(chan float64)
		in[channelNumber] <- Data{
			Account: account,
			Result:  &result,
		}
		select {
		case result := <-result:
			log.Printf("account: %v, result: %+v\n", account, result)
			wg.Done()
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
	in = make([]chan Data, maxThread)

	session, _ := mgo.Dial("localhost:27017")
	globalDB = session.DB("logs")

	globalDB.C("bank").DropCollection()

	for i := range in {
		in[i] = make(chan Data)
	}

	// create 100 user
	for i := 0; i < maxUser; i++ {
		account := "user" + strconv.Itoa(i+1)
		user := currency{Account: account, Amount: 1000.00, Code: "USD", Version: 1}
		if err := globalDB.C("bank").Insert(&user); err != nil {
			panic("insert error")
		}
	}

	for i := range in {
		go func(in *chan Data, i int) {
			for {
				select {
				case data := <-*in:
				LOOP:
					entry := currency{}
					// step 1: get current amount
					err := globalDB.C("bank").Find(bson.M{"account": data.Account}).One(&entry)

					if err != nil {
						panic(err)
					}

					//step 3: subtract current balance and update back to database
					entry.Amount = entry.Amount + 50.000
					err = globalDB.C("bank").Update(bson.M{
						"version": entry.Version,
						"_id":     entry.ID,
					}, bson.M{"$set": map[string]interface{}{
						"amount":  entry.Amount,
						"version": (entry.Version + 1),
					}})

					if err != nil {
						log.Println("got errors: ", err)
						goto LOOP
					}

					*data.Result <- entry.Amount
				}
			}

		}(&in[i], i)
	}

	log.Println("Listen server on " + port + " port")
	http.HandleFunc("/", pay)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
