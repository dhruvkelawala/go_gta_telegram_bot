package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)


type Rockstar struct {
	ID int `json:"id"`
	Name string `json:"name"`
	StatusCode int `json:"status"`
	StatusTag string `json:"status_tag"`
	RecentUpdate string `json:"recent_update"`
}

type RockstarStatuses struct{
	Statuses []Rockstar
}



func checkStatus() Rockstar {

	//var r string
	response, err := http.Get("https://support.rockstargames.com/services/status.json?tz=Asia/Calcutta")

	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	var rockstar RockstarStatuses

	err = json.Unmarshal(body, &rockstar)

	if err!= nil{
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	for _, status := range rockstar.Statuses {

			if status.ID == 3 {
				return status
			}
	}

	return Rockstar{}
}
