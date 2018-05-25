package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var globalDB *mgo.Database
var account = "appleboy"

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
LOOP:
	entry := currency{}
	// step 1: get current amount
	err := globalDB.C("bank").Find(bson.M{"account": account}).One(&entry)

	if err != nil {
		panic(err)
	}

	wait := Random(1, 100)
	time.Sleep(time.Duration(wait) * time.Millisecond)

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
		goto LOOP
	}

	fmt.Printf("%+v\n", entry)

	io.WriteString(w, "ok")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	session, _ := mgo.Dial("localhost:27017")
	globalDB = session.DB("queue")
	globalDB.C("bank").DropCollection()

	user := currency{Account: account, Amount: 1000.00, Code: "USD", Version: 1}
	err := globalDB.C("bank").Insert(&user)

	if err != nil {
		panic("insert error")
	}

	log.Println("Listen server on " + port + " port")
	http.HandleFunc("/", pay)
	http.ListenAndServe(":"+port, nil)
}
