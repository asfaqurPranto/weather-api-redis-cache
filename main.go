package main

import (
	"context"
	"os"
	//"encoding/json"
	"fmt"
	"net/http"
	"time"

	"io"

	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

//Adding redis caching
var ctx =context.Background()
var rdx_client *redis.Client
func initRedis(){
	rdx_client=redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR") ,
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB: 0,  // use default DB
	})
}
// func resetCache() error {
//     return rdx_client.FlushAll(ctx).Err()
// }
func initEnv(){
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}
func GetCachedWeather(City string) (string,error){
	val,err:=rdx_client.Get(ctx,City).Result()
	if err==redis.Nil{
		return "",redis.Nil
	}

    return val,err

}
func setCachedWeather(City string, x string){
	//data,_:=json.Marshal(x)
	rdx_client.Set(ctx,City,x,2*time.Hour)
}


func WeatherHandler(w http.ResponseWriter, r *http.Request) {


	url_var := mux.Vars(r)
	City := url_var["city"]
	result,err:=GetCachedWeather(City)

	if err==redis.Nil{   //cached missed
		fmt.Println("Cached missed")
		API_KEY := os.Getenv("API_KEY")
		url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", City, API_KEY)

		resp, err := http.Get(url)
		if err != nil {
			http.Error(w,err.Error(),404)
			return
		}
		
		defer resp.Body.Close()

		body,_:=io.ReadAll(resp.Body)// //[]byte to json string
		result=string(body)
		setCachedWeather(City,result)

	}else if err!=nil{
		http.Error(w,err.Error(),404)
		return
	}
						
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))

}



func main() {

	mux := mux.NewRouter()
	
	initRedis()
	//resetCache()
	initEnv()
	mux.HandleFunc("/weather/{city}", WeatherHandler).Methods("GET")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Print(err)
	}

}
