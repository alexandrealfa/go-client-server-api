package main

import (
	"context"
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

func getURL() (url string) {
	url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	return
}

type Price struct {
	url string
}

type LogSchema struct {
	Id       int `gorm:"primaryKey"`
	Value    string
	CoinType string
}

type serializedJson struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResponseSchema struct {
	USDBRL struct {
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

func (p Price) ServeHTTP(w http.ResponseWriter, _ *http.Request) {

	price, err := p.getPrice()
	if err != nil {
		log.Println("Error to get Price: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dataLog := LogSchema{
		Value:    price.USDBRL.Bid,
		CoinType: price.USDBRL.Code,
	}

	p.logData(dataLog)

	response := serializedJson{"Dólar", price.USDBRL.Bid}

	responseJson, marshalError := json.Marshal(response)
	if marshalError != nil {
		log.Println("error to Marshal response: ", marshalError.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(responseJson); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal("error to write response: ", err.Error())
	}
}

func (p Price) getPrice() (ResponseSchema, error) {
	var data ResponseSchema
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	requestCtx, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		log.Println("Error in request process: ", err.Error())

		return data, err
	}

	req, err := http.DefaultClient.Do(requestCtx)
	if err != nil {
		log.Println("Error to get request data: ", err.Error())
		return data, err
	}

	res, parseError := io.ReadAll(req.Body)

	if parseError != nil {
		log.Println("error to read response: ", res)
		return data, parseError
	}

	if MarshalError := json.Unmarshal(res, &data); MarshalError != nil {
		log.Println("error in Json parser process: ", MarshalError.Error())
		return data, MarshalError
	}

	return data, err
}

func (p Price) logData(data LogSchema) {
	var (
		dsn                       = "root:root@tcp(localhost:3306)/client-server-db?charset=utf8mb4&parseTime=True&loc=Local"
		timeoutTime time.Duration = 20
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("error to connect with database: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*timeoutTime)
	defer cancel()
	db.WithContext(ctx)

	if err := db.AutoMigrate(&LogSchema{}); err != nil {
		log.Println("Error to create table: ", err)
	}

	db.Create(&data)
}

func main() {
	mutex := http.NewServeMux()
	mutex.Handle("/cotacao", Price{getURL()})

	if err := http.ListenAndServe(":8080", mutex); err != nil {
		log.Fatal("error to run server: ", err.Error())
	}
}
