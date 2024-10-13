package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"github.com/gorilla/mux"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"encoding/json"
	"log"
	"time"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"github.com/golang-jwt/jwt/v5"
)

var ctx = context.Background()

type DB struct {
	Username string `json:"database_user"`
	Password string `json:"database_password"`
	Host     string `json:"database_host"`
	DBName   string `json:"database_name"`
}

type User struct {
	ID string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Oauth int `json:"oauth"`
}

type Register struct {
	Username string `json:"username"`
	Password string `json:"password"`
	VerifyPassword string `json:"verify_password"`
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Oauth struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthorizeUrl string `json:"authorize_url"`
	TokenUrl string `json:"token_url"`
	Scopes	string `json:"scopes"`
	UserInfo string `json:"user_info"`
}

type JwtSecret struct {
	JwtSecretKey string `json:"jwt_secret_key"`
}

// Helper Functions 
// Get secret from AWS Secrets Manager
func GetSecret(secretName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return "", err
	}

	client := secretsmanager.NewFromConfig(cfg)
	secretValue, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})

	if err != nil {
		return "", err
	}

	if secretValue.SecretString != nil {
		return *secretValue.SecretString, nil
	}

	return "", nil
}

func GetDBConnectionString() (string, error) {
	secret, err := GetSecret("dev/ls/db")
	if err != nil {
		fmt.Println(err)
	}

	var db DB 
	err = json.Unmarshal([]byte(secret), &db)
	if err != nil {
		fmt.Println(err)
	}

	return fmt.Sprintf("%s:%s@tcp(%s)/%s", db.Username, db.Password, db.Host, db.DBName), nil
}

func GetOauthConfig() (Oauth, error) {
	secret, err := GetSecret("dev/ls/oauth")
	if err != nil {
		fmt.Println(err)
	}

	var oauth Oauth
	err = json.Unmarshal([]byte(secret), &oauth)
	if err != nil {
		fmt.Println(err)
	}

	return oauth, nil
}

func CheckPasswordComplexity(password string) bool {
	if len(password) < 8 {
		return false
	}
	// Check for at least one uppercase letter
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return false
	}
	// Check for at least one lowercase letter
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return false
	}
	// Check for at least one number
	if !strings.ContainsAny(password, "0123456789") {
		return false
	}
	// Check for at least one special character
	if !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		return false
	}
	
	return true

}

func HashPassword(password string) (string, string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	return string(hash), ""
}

func VerifyPassword(hashedPass string, plainPassword []byte) bool {
	
	byteHash := []byte(hashedPass)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	if err != nil {
		return false
	}
	return true
}

func CheckIfUserExists(username string) bool {
	DbConn, err := GetDBConnectionString()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", DbConn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	query := "SELECT username FROM users WHERE username = ?"
	row := db.QueryRow(query, username).Scan(&username)
	if row != nil {
		if row != sql.ErrNoRows {
			log.Fatal(err)
		}
		return false
	}
	return true
}

func CreateJwtToken(username, id string) (string, error) {
	secret, err := GetSecret("dev/ls/jwt")
	if err != nil {
		return "", err
	}

	var jwtSecret JwtSecret
	err = json.Unmarshal([]byte(secret), &jwtSecret)
	if err != nil {
		return "", err
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"id": id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iss": "auth-server",
		"iat": time.Now().Unix(),
	})

	token, err := claims.SignedString([]byte(jwtSecret.JwtSecretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}


// Route Functions 
func OauthAuthorize(w http.ResponseWriter, r *http.Request) {}

func OauthCallback(w http.ResponseWriter, r *http.Request) {}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	DbConn, err := GetDBConnectionString()
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get database connection string", http.StatusInternalServerError)
	}
	
	db, err := sql.Open("mysql", DbConn)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not open a connection to the database", http.StatusInternalServerError)
	}

	defer db.Close()

	reqBody, _ := io.ReadAll(r.Body)
	var register Register
	json.Unmarshal(reqBody, &register)

	if register.Password != register.VerifyPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	if !CheckPasswordComplexity(register.Password) {
		http.Error(w, "Password does not meet complexity requirements", http.StatusBadRequest)
		return
	}

	if CheckIfUserExists(register.Username) {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	hashedPass, _ := HashPassword(register.Password)

	query := "INSERT INTO users (username, password) VALUES (?, ?)"
	_, err = db.Exec(query, register.Username, hashedPass)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not insert user into the database", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})

	return
}

func Login(w http.ResponseWriter, r *http.Request) {
	DbConn, err := GetDBConnectionString()
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get database connection string", http.StatusInternalServerError)
	}
	
	db, err := sql.Open("mysql", DbConn)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not open a connection to the database", http.StatusInternalServerError)
	}

	defer db.Close()

	reqBody, _ := io.ReadAll(r.Body)

	var user LoginUser
	json.Unmarshal(reqBody, &user)

	var dbUser User
	query := "SELECT id, username, password FROM users WHERE username = ?"
	row := db.QueryRow(query, user.Username).Scan(&dbUser.ID, &dbUser.Username, &dbUser.Password)
	if row != nil {
		if row == sql.ErrNoRows {
			http.Error(w, "User does not exist", http.StatusBadRequest)
			return
		}
		log.Fatal(row)
		http.Error(w, "Could not query the database", http.StatusInternalServerError)
	}

	if !VerifyPassword(dbUser.Password, []byte(user.Password)) {
		http.Error(w, "Password is incorrect", http.StatusBadRequest)
		return
	}

	token, err := CreateJwtToken(dbUser.Username, dbUser.ID)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not create JWT token", http.StatusInternalServerError)
	}



	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token, "message": "User logged in successfully"})
	return
}

func VerifyToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "No token provided", http.StatusBadRequest)
		return
	}

	secret, err := GetSecret("dev/ls/jwt")
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get JWT secret", http.StatusInternalServerError)
		return
	}

	var jwtSecret JwtSecret
	err = json.Unmarshal([]byte(secret), &jwtSecret)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not unmarshal JWT secret", http.StatusInternalServerError)
		return
	}

	_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret.JwtSecretKey), nil
	})

	if err != nil {
		http.Error(w, "Token is invalid", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Token is valid"})
	return
}



func handleRequests() {
	auth_router := mux.NewRouter().StrictSlash(true)
	auth_router.HandleFunc("/proxy/v1/auth/oauth/authorize/{provider}", OauthAuthorize).Methods("GET")
	auth_router.HandleFunc("/proxy/v1/auth/oauth/callback/{provider}", OauthCallback).Methods("GET")
	auth_router.HandleFunc("/proxy/v1/auth/register", RegisterUser).Methods("POST")
	auth_router.HandleFunc("/proxy/v1/auth/login", Login).Methods("POST")
	auth_router.HandleFunc("/proxy/v1/auth/verify", VerifyToken).Methods("GET")

	// Fix Cors
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(auth_router)
	http.ListenAndServe(":8080", handler)
}

func main() {
	fmt.Println("Starting Auth Server")
	handleRequests()
}