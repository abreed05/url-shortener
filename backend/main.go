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
	"github.com/redis/go-redis/v9" 
	"encoding/json"
	"math/rand"
	"log"
	"time"
)

var ctx = context.Background()

type DB struct {
	Username string `json:"database_user"`
	Password string `json:"database_password"`
	Host     string `json:"database_host"`
	DBName   string `json:"database_name"`
}

type URL struct {
	ID string `json:"id"`
	ShortURL string `json:"short_url"`
	LongURL string `json:"long_url"`
	UserID string `json:"user_id"`
}

type PostURL struct {
	LongURL string `json:"long_url"`
}

// Helper Functions

func RedisClient() *redis.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic(err)
	}

	SecretClient := secretsmanager.NewFromConfig(cfg)
	secretName := "dev/ls/db"
	secretValue, err := SecretClient.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})

	if err != nil {
		panic(err)
	}

	var secret DB
	err = json.Unmarshal([]byte(*secretValue.SecretString), &secret)
	if err != nil {
		panic(err)
	}

	connectionString := fmt.Sprintf("%s:%s", secret.Host, "6379")

	client := redis.NewClient(&redis.Options{
		Addr:     connectionString,
		Password: "",
		DB:	   0,
		Protocol: 3,
	})

	return client
}

// Get Value from redis by key 
func GetValue(key string) (string, error) {
	ctx := context.Background()
	client := RedisClient()
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

// Set Value in redis by key
func SetValue(key string, value string) error {
	ctx := context.Background()
	client := RedisClient()
	_, err := client.Set(ctx, key, value, 24 * time.Hour).Result()
	if err != nil {
		return err
	}

	return nil
}

func UpdateCacheExpiry(key string) error {
	ctx := context.Background()
	client := RedisClient()
	_, err := client.Expire(ctx, key, 24 * time.Hour).Result()
	if err != nil {
		return err
	}

	return nil
}


// get secret from secrets manager 
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

	var secret DB
	err = json.Unmarshal([]byte(*secretValue.SecretString), &secret)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s@tcp(%s)/%s", secret.Username, secret.Password, secret.Host, secret.DBName), nil
}

func generateShortURL() (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortURL := make([]byte, 6)
	for i := range shortURL {
		shortURL[i] = chars[rand.Intn(len(chars))]
	}

	return string(shortURL), nil
}

// Route Functions

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	SecretName := "dev/ls/db"
	secret, err := GetSecret(SecretName)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get secret from Secrets Manager", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("mysql", secret)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not open a connection to the DB", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	shortUrlCode, err := generateShortURL()
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not generate short code", http.StatusInternalServerError)
		return
	}
	
	reqBody, _ := io.ReadAll(r.Body)
	var postURL PostURL
	json.Unmarshal(reqBody, &postURL)

	// check if shorturl already exists in the DB 
	result, err := db.Query("SELECT COUNT(*) FROM link_shortener WHERE short_url = ?", shortUrlCode)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get key from DB", http.StatusInternalServerError)
		return
	}

	var count int
	for result.Next() {
		result.Scan(&count)
	}

	// while loop to check if the count is actually 0, if not it attempts to generate another unique code
	for {
		if count == 0 {
			break;
		}

		shortUrlCode, err = generateShortURL()
		if err != nil {
			log.Fatal(err)
			http.Error(w, "Could not generate short code", http.StatusInternalServerError)
			return
		}

		result, err = db.Query("SELECT COUNT(*) FROM link_shortener WHERE short_url = ?", shortUrlCode)
		if err != nil {
			log.Fatal(err)
			http.Error(w, "Could not get key from DB", http.StatusInternalServerError)
			return
		}

		for result.Next() {
			result.Scan(&count)
		}

	}

	// insert the new shorturl into the DB
	_, err = db.Exec("INSERT INTO link_shortener (short_url, long_url) VALUES (?, ?)", shortUrlCode, postURL.LongURL)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not insert into DB table", http.StatusInternalServerError)
		return
	}

	// set the shorturl in the cache
	err = SetValue(shortUrlCode, postURL.LongURL)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not set value in cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"short_url": shortUrlCode})

	return
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	// check if the shorturl exists in the cache
	val, err := GetValue(shortURL)
	if err == nil {
		// update the cache expiry
		err = UpdateCacheExpiry(shortURL)
		http.Redirect(w, r, val, http.StatusMovedPermanently)
		return
	}
	
	SecretName := "dev/ls/db"
	secret, err := GetSecret(SecretName)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get secret from Secrets Manager", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("mysql", secret)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not open a connection to the DB", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	result, err := db.Query("SELECT long_url FROM link_shortener WHERE short_url = ?", shortURL)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not get key from DB", http.StatusInternalServerError)
		return
	}

	var longURL string
	for result.Next() {
		result.Scan(&longURL)
	}

	// set the shorturl in the cache
	err = SetValue(shortURL, longURL)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Could not set value in cache", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
	return

}


// Handle requests and initialize the main function 
func handleRequests() {
	url_router := mux.NewRouter().StrictSlash(true)
	url_router.HandleFunc("/proxy/v1/shorten-url", ShortenURL).Methods("POST")
	url_router.HandleFunc("/{shortURL}", RedirectURL).Methods("GET")
	http.ListenAndServe(":8080", url_router)
}

func main() {
	fmt.Println("Starting the URL Shortener service...")
	handleRequests()
}