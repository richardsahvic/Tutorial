package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"service"

	"datasource"
	"repo"
	"request"

	"github.com/bwmarrin/snowflake"
)

var db = datasource.InitConnection()
var userService = service.NewUserService(repo.NewRepository(db))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 5000))

	var loginReq request.LoginRequest
	err := json.Unmarshal(body, &loginReq)
	if err != nil {
		log.Println("ERROR at unmarshal", err)
		return
	}

	loginResult, err := userService.Login(loginReq.Username, loginReq.Password, loginReq.Role)
	if err != nil {
		log.Println("Failed at login,   ", err)
	}

	var loginResp request.Response

	if len(loginResult) == 0 {
		loginResp.Message = "Login failed"
	} else {
		loginResp.Message = "Login Success"
	}

	js, err := json.Marshal(loginResp)
	if err != nil {
		log.Println("ERROR at login marshal,    ", err)
	}

	w.Header().Set("token", loginResult)
	w.Write(js)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 5000))

	var regRequest request.RegisterRequest
	json.Unmarshal(body, &regRequest)

	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println("Fail to generate snowflake id,    ", err)
		return
	}

	id := node.Generate().String()

	userRegister := repo.User{
		ID:       id,
		Email:    regRequest.Email,
		Msisdn:   regRequest.Msisdn,
		Username: regRequest.Username,
		Password: regRequest.Password,
		Status:   0,
	}

	role := regRequest.Role

	registerResult, err := userService.Register(userRegister, role)
	if err != nil {
		log.Println("failed to register,    ", err)
	}

	var regResponse request.Response

	if !registerResult {
		regResponse.Message = "Register failed"
	} else {
		regResponse.Message = "Register success"
	}

	json.NewEncoder(w).Encode(regResponse)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenHeader := r.Header.Get("token")

	profile, err := userService.ViewProfile(tokenHeader)
	if err != nil {
		log.Println("Failed to view profile,    ", err)
	}

	profile.Password = "*"

	json.NewEncoder(w).Encode(profile)
}