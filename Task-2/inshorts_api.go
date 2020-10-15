package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"sync"
)

var mutex = &sync.Mutex{} // mutual-exclusion lock for achieving process synchronization

var curr_id = 3 //As already in main function there are two articles I have added by default, so curr_id will start from 3

//var result map[string]interface{} -Declared an empty map interface

type Article struct {
	ID        string `json:"ID"`
	Title     string `json:"Title"`
	SubTitle  string `json:"SubTitle"`
	Content   string `json:"Content"`
	Timestamp string `json:"Timestamp"`
}

// Articles ... global array
// which we can populate later
var Articles []Article

//Create a Article and List all articles - 1st & 3rd endpoint
func allArticles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":

		//Here we are using offset pagination
		page, flag1 := r.URL.Query()["page"] 
		limit, flag2 := r.URL.Query()["limit"]
		if !flag1 || !flag2 || len(page[0])<1 || len(limit[0])<1{
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Articles)	
			return
		}
		pageNo,e1 := strconv.Atoi(page[0]) //page is a list in which the first entry is page number
		maxLimit,e2 := strconv.Atoi(limit[0]) //max articles per page
			
		//If there is an invalid parameter for page we return all of the articles
		if e1!=nil || e2!=nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Articles)	
				
		}else{
			//Indexing starts from 0
			if pageNo*maxLimit >= len(Articles){
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w,"Not Found")
			}else{
				w.WriteHeader(http.StatusOK)
				for i:=pageNo*maxLimit; i<min(len(Articles),(pageNo+1)*maxLimit);i++{
					json.NewEncoder(w).Encode(Articles[i])
				}
			}
		}

	case "POST":
		//Critical section where a potential race condition may occur
		mutex.Lock()
		w.WriteHeader(http.StatusCreated)
		var a Article
		// Lets decode the request body and pass the error
		// If it is not nill then consider it as a bad request
		e := json.NewDecoder(r.Body).Decode(&a)
		a.ID = curr_id
		a.Timestamp = time.Now().Format(time.RFC850)
		if e == nil {
			curr_id++
			Articles = append(Articles, a)
		} else {
			http.Error(w, e.Error(), http.StatusBadRequest)
		}
		mutex.Unlock()
	}
}


//Fetching article using id - 2nd Endpoint
func articleByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := strings.TrimPrefix(r.URL.Path, "/articles/")
	if len(id) == 0 {
		articles(w, r)
	} else {
		var flag bool
		flag = false
		for _, article := range Articles {
			if article.ID == id {
				flag = true
				json.NewEncoder(w).Encode(article)
			}
		}
		if !flag {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not Found")
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w,"Article found")
		}
	}
}

//Search an article - 4th endpoint
func searchArticles(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["q"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	key := keys[0]
	var flag bool
	flag = false
	for i,article := range Articles
	{
		if article.Title == key || article.SubTitle == key || article.Content == key {
			flag = true
			json.NewEncoder(w).Encode(article)
		}
	}
	if flag {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w,"Found")
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not Found")
	}
}


//Handling all the requests and routing them to different urls
func handleRequests() {
	http.HandleFunc("/articles", allArticles)
	http.HandleFunc("/articles/", articleByID)
	http.HandleFunc("/articles/search", searchArticles)
	//Listening to port - 3000 on our localhost
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func main() {
	Articles = []Article{
		Article{ID: "1", Title: "Sample title 1", SubTitle: "subtitle 1", Content: "Content 1", Timestamp: time.Now().Format(time.RFC850)},
		Article{ID: "2", Title: "Sample title 2", SubTitle: "subtitle 2", Content: "Content 2", Timestamp: time.Now().Format(time.RFC850)},
	}
	
	handleRequests()
}
