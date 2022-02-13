package main

// 2HUKB03MKDX4VKXV
// https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=IBM&interval=5min&apikey=demo

//
//https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=IBM&interval=5min&apikey=2HUKB03MKDX4VKXV

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/quote"
)

func (a *API)getDataHandler(c echo.Context) error{
	data ,cacheHit,err := a.getData(c.Request().Context(),c.Param("stock_name"))
	 if err != nil{
	 	log.Printf("%s\n",err)
	 	c.Response().Writer.WriteHeader(http.StatusInternalServerError)
	 }

	resp := APIResponse{
		Cache: cacheHit,
		Data: data,
	}
	if err != nil{
		log.Printf("%s\n",err)
		c.Response().Writer.WriteHeader(http.StatusInternalServerError)
	}
		return c.JSON(http.StatusOK,resp)
}
func (a *API)getData(ctx context.Context,stockName string)(*finance.Quote,bool,error){
	// checking if the query cached
	value,err := a.cache.Get(ctx,stockName).Result()
	if err == redis.Nil{
		// call external data source
	q,err := quote.Get(stockName)
	if err != nil{
		return nil,false,err
	}

	b,err := json.Marshal(q)
	if err != nil{
		return nil,false,err
	}
		//set the value
		err = a.cache.Set(ctx,stockName,bytes.NewBuffer(b).Bytes(),time.Second*15).Err()
	  if err != nil{
		return nil,false,err
	  }

		//return the response 
	return q,false,nil

	}else if err != nil {
		return nil,false,err
	}else{
		var data *finance.Quote
		err := json.Unmarshal(bytes.NewBufferString(value).Bytes(),&data)
	if err != nil{
		return nil,false,err
	}


	return data,true,nil
	}

}

// main function 
func main() {
	api := NewAPI()
	e := echo.New()
	e.GET("/stock/:stock_name", api.getDataHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
// API struct
type API struct {
	cache *redis.Client

}

func NewAPI() *API{
	// get redis address from docker env  var
	//redisAddress := fmt.Sprintf("%s:6379",os.Getenv("REDIS_URL"))
	 rdb := redis.NewClient(&redis.Options{
        Addr:    "localhost:6379", 
        Password: "", // no password set
        DB:       0,  // use default DB
    })
		return &API{
			cache: rdb,
		}
}

type APIResponse struct {
	Cache bool `json:"cache"`
	Data *finance.Quote `json:"data"`
}
	
