package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CurrencyValue struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "currency.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS quotes (timestamp DATETIME NOT NULL, bid FLOAT NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		req, err := http.NewRequest("GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}

		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "failed to fetch currency quote", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "unexpected status code", http.StatusInternalServerError)
			return
		}

		var quote CurrencyValue
		err = json.NewDecoder(resp.Body).Decode(&quote)
		if err != nil {
			http.Error(w, "failed to parse currency quote", http.StatusInternalServerError)
			return
		}

		stmt, err := db.Prepare("INSERT INTO quotes (timestamp, bid) VALUES (?, ?)")
		if err != nil {
			http.Error(w, "failed to prepare SQL statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		ctx, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		_, err = stmt.Exec(time.Now(), quote.Usdbrl.Bid)
		if err != nil {
			fmt.Println(err)

			http.Error(w, "failed to persist currency quote", http.StatusInternalServerError)
			return
		}

		log.Println(quote)
		json.NewEncoder(w).Encode(quote)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
