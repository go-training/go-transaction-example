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
var mu = &sync.Mutex{}

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
	entry := currency{}
	//Solution here, Lock other thread access this section of code until it's unlocked
	mu.Lock()
	defer mu.Unlock()
	// step 1: get current amount
	err := globalDB.C("bank").Find(bson.M{"account": account}).One(&entry)

	if err != nil {
		panic(err)
	}

	wait := Random(1, 100)
	time.Sleep(time.Duration(wait) * time.Millisecond)

	//step 3: subtract current balance and update back to database
	entry.Amount = entry.Amount + 50.000
	err = globalDB.C("bank").UpdateId(entry.ID, &entry)

	if err != nil {
		panic("update error")
	}

	fmt.Printf("%+v\n", entry)

	io.WriteString(w, "ok")
}

func main() {
	session, _ := mgo.Dial("localhost:27017")
	globalDB = session.DB("logs")

	globalDB.C("bank").DropCollection()

	user := currency{Account: account, Amount: 1000.00, Code: "USD"}
	err := globalDB.C("bank").Insert(&user)

	if err != nil {
		panic("insert error")
	}

	log.Println("Listen server on 8000 port")
	http.HandleFunc("/", pay)
	http.ListenAndServe(":8000", nil)
}
