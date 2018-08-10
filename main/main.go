package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"repo"
	"service"

	"github.com/bwmarrin/snowflake"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type LoginRequest struct {
	Username string `json: "username"`
	Password string `json: "password"`
	Role     int    `json: "role"`
}

type Response struct {
	Token   string `json:"token"`
	Message string `json: "message"`
}

type RegisterRequest struct {
	ID       string `json: "id"`
	Email    string `json: "email"`
	Msisdn   string `json: "msisdn"`
	Username string `json: "username"`
	Password string `json: "password"`
	Status   int    `json: "status"`
	Role     int    `json: "role"`
}

type ProfileRequest struct {
	Token string `json: "token"`
	Email string `json: "email"`
}

var userService service.UserService

func main() {
	db, err := sqlx.Connect("mysql", "dev:dev@(localhost:3306)/myapp?parseTime=true")
	if err != nil {
		log.Fatalln("Failed to connect to database,    ", err)
	}

	r := repo.NewRepository(db)

	userService = service.NewUserService(r)

	route := mux.NewRouter()
	route.HandleFunc("/login", loginHandler).Methods("POST")
	route.HandleFunc("/register", registerHandler).Methods("POST")
	route.HandleFunc("/viewprofile", profileHandler).Methods("POST")

	http.Handle("/", route)
	log.Println("SERVER STARTED")

	http.ListenAndServe(":8080", route)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 5000))

	var loginRequest LoginRequest
	json.Unmarshal(body, &loginRequest)

	loginResult, err := userService.Login(loginRequest.Username, loginRequest.Password, loginRequest.Role)
	if err != nil {
		log.Println("Failed at login,   ", err)
	}

	var loginResponse Response

	if len(loginResult) == 0 {
		loginResponse.Message = "Login failed"
	} else {
		loginResponse = Response{
			Token:   loginResult,
			Message: "Login Success",
		}
	}

	json.NewEncoder(w).Encode(loginResponse)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 5000))

	var registerRequest RegisterRequest
	json.Unmarshal(body, &registerRequest)

	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println("Fail to generate snowflake id,    ", err)
		return
	}

	id := node.Generate().String()

	userRegister := repo.User{
		ID:       id,
		Email:    registerRequest.Email,
		Msisdn:   registerRequest.Msisdn,
		Username: registerRequest.Username,
		Password: registerRequest.Password,
		Status:   0,
	}

	role := registerRequest.Role

	registerResult, err := userService.Register(userRegister, role)
	if err != nil {
		log.Println("failed to register,    ", err)
	}

	var registerResponse Response

	if !registerResult {
		registerResponse.Message = "Register failed"
	} else {
		registerResponse.Message = "Register success"
	}

	json.NewEncoder(w).Encode(registerResponse)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenHeader := r.Header.Get("token")
	emailHeader := r.Header.Get("email")

	var profileRequest ProfileRequest
	profileRequest.Token = tokenHeader
	profileRequest.Email = emailHeader

	profile, err := userService.ViewProfile(emailHeader, tokenHeader)
	if err != nil {
		log.Println("Failed to view profile,    ", err)
	}

	_, err = json.Marshal(profile)
	if err != nil {
		log.Println("ERROR profile marshal,    ", err)
	}

	w.Header().Set("id", profile.ID)
	w.Header().Set("email", profile.Email)
	w.Header().Set("msisdn", profile.Msisdn)
	w.Header().Set("username", profile.Username)
}
