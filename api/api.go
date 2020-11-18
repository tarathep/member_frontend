package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/tarathep/member_frontend/model"
)

type MyJsonName struct {
	Members []model.Member `json:"members"`
}

func GetMembers() []model.Member {
	resp, err := http.Get("http://localhost:8080/members")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data := MyJsonName{}
	json.Unmarshal([]byte(string(body)), &data)
	return data.Members
}

func AddMembers(member model.Member) {

	reqBody, err := json.Marshal(member)
	if err != nil {
		log.Fatalln(err)
	}

	request, err := http.NewRequest("POST", "http://localhost:8080/members", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalln(err)
	}

	client := http.Client{Timeout: time.Duration(5 * time.Second)}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(member)

	data := model.Message{}
	json.Unmarshal([]byte(string(body)), &data)
	fmt.Println(data)
}

func EditMembers(member model.Member) {

	reqBody, err := json.Marshal(member)
	if err != nil {
		log.Fatalln(err)
	}

	request, err := http.NewRequest("PUT", "http://localhost:8080/members", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalln(err)
	}

	client := http.Client{Timeout: time.Duration(5 * time.Second)}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(member)

	data := model.Message{}
	json.Unmarshal([]byte(string(body)), &data)
	fmt.Println(data)
}

func DeleteMembers(id string) {

	request, err := http.NewRequest("DELETE", "http://localhost:8080/members/"+id, nil)
	if err != nil {
		log.Fatalln(err)
	}

	client := http.Client{Timeout: time.Duration(5 * time.Second)}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data := model.Member{}
	json.Unmarshal([]byte(string(body)), &data)
}

func Login(username string, password string) model.Auth {

	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data := model.Auth{}
	json.Unmarshal([]byte(string(body)), &data)
	return data
}
