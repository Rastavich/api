package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Company struct {
	Id          int
	Account     string
	Sys         string
	Username    string
	Pword       string
	Description string
	Address     string
	Grouping    string
	Notes       string
}

const (
	dbhost = "DB_HOST"
	dbport = "DB_PORT"
	dbuser = "DB_USER"
	dbpass = "DB_PASS"
	dbname = "DB_NAME"
)

func main() {
	initDb()
	defer db.Close()
	http.HandleFunc("/api/getAllCompanies", GETAllCompanies)
	http.HandleFunc("/api/getCompanyByName/", GETCompanyByName)
	http.HandleFunc("/api/updateField/", UPDATEfield)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func GETAllCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rows, err := db.Query(`
            SELECT * 
            FROM public."data" `)

	if err != nil {
		panic(err)
	}

	var company []Company

	for rows.Next() {
		var client Company
		rows.Scan(&client.Id, &client.Account, &client.Sys, &client.Username,
			&client.Pword, &client.Description, &client.Address,
			&client.Grouping, &client.Notes)
		company = append(company, client)
	}

	companyBytes, _ := json.MarshalIndent(company, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(companyBytes)

	defer rows.Close()
}

// Go function to get company bu account name
func GETCompanyByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	account := r.URL.Query().Get("account")
	rows, err := db.Query("SELECT * FROM public.\"data\"")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var company []Company

	for rows.Next() {

		var client Company
		if err := rows.Scan(&client.Id, &client.Account, &client.Sys, &client.Username,
			&client.Pword, &client.Description, &client.Address,
			&client.Grouping, &client.Notes); err != nil {

			if strings.Contains(strings.ToLower(client.Account), strings.ToLower(account)) {
				company = append(company, client)
			}

		}
	}

	companyBytes, _ := json.MarshalIndent(company, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(companyBytes)

}

func UPDATEfield(w http.ResponseWriter, r *http.Request) {
	// update a field in the database
	w.Header().Set("Access-Control-Allow-Origin", "*")
	id := r.URL.Query().Get("id")
	field := r.URL.Query().Get("field")
	value := r.URL.Query().Get("value")

	_, err := db.Exec(`
		UPDATE public."data"	
		SET $1 = $2
		WHERE account = $3
		`, field, value, id)

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully updated!")

	defer db.Close()
}

func initDb() {
	config := dbConfig()
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport],
		config[dbuser], config[dbpass], config[dbname])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
}

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := "localhost", true
	if !ok {
		panic("DBHOST environment variable required but not set")
	}
	port, ok := "5432", true
	if !ok {
		panic("DBPORT environment variable required but not set")
	}
	user, ok := "postgres", true
	if !ok {
		panic("DBUSER environment variable required but not set")
	}
	password, ok := "password", true
	if !ok {
		panic("DBPASS environment variable required but not set")
	}
	name, ok := "customerLogins", true
	if !ok {
		panic("DBNAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}
