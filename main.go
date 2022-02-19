package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"encoding/csv"
	_ "github.com/lib/pq"

	"github.com/rs/cors"
)

var db *sql.DB

const (
	dbhost = "DB_HOST"
	dbport = "DB_PORT"
	dbuser = "DB_USER"
	dbpass = "DB_PASS"
	dbname = "DB_NAME"
)

// Seed type
type Seed struct {
	db *sql.DB
}

type Company struct {
	Id             string
	Account        string
	Sys            string
	Username       string
	Pword          string
	Description    string
	Address        string
	Grouping       string
	Notes          string
	Aka            string
	Account_status string
}

func (s Seed) CustomerSeed() {
	fmt.Println("Starting Customer Seed")
	csvFile, err := os.Open("Clients3.csv")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()

	if err != nil {
		fmt.Println(err)
	}
	for _, line := range csvLines {
		cust := Company{
			Id:             line[0],
			Account:        line[1],
			Sys:            line[2],
			Username:       line[3],
			Pword:          line[4],
			Description:    line[5],
			Address:        line[6],
			Grouping:       line[7],
			Notes:          line[8],
			Aka:            line[9],
			Account_status: line[10],
		}
		fmt.Println(cust.Id + " " + cust.Account + " " + cust.Sys + " " + cust.Username + " " + cust.Pword + " " + cust.Description + " " + cust.Address + " " + cust.Grouping + " " + cust.Notes + " " + cust.Aka + " " + cust.Account_status)

		//prepare the statement
		stmt, _ := s.db.Prepare(`INSERT INTO customers(Id, account, aka, username, pword, sys, description, address, grouping, notes, account_status) VALUES (?,?,?,?,?,?,?,?,?,?,?)`)
		// execute query
		_, err := stmt.Exec(cust.Id, cust.Account, cust.Aka, cust.Username, cust.Pword, cust.Sys, cust.Description, cust.Address, cust.Grouping, cust.Notes, cust.Account_status)
		if err != nil {
			panic(err)
		}

	}
}

// Execute will executes the given seeder method
func Execute(db *sql.DB, seedMethodNames ...string) {
	s := Seed{db}

	log.Println("Seeding...", seedMethodNames)

	// Execute only the given method names
	for _, item := range seedMethodNames {
		seed(s, item)
	}
}

func seed(s Seed, seedMethodName string) {
	// Get the reflect value of the method
	m := reflect.ValueOf(s).MethodByName(seedMethodName)
	// Exit if the method doesn't exist
	if !m.IsValid() {
		log.Fatal("No method called ", seedMethodName)
	}
	// Execute the method
	log.Println("Seeding", seedMethodName, "...")
	m.Call(nil)
	log.Println("Seed", seedMethodName, "succeed")
}



func main() {
	handleArgs()
	defer db.Close()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost/:1"},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/getAllCompanies", GETAllCompanies)
	mux.HandleFunc("/api/getCompanyByName/", GETCompanyByName)
	mux.HandleFunc("/api/updateField/", UPDATEfield)
	mux.HandleFunc("/api/deleteCompanyRowById/", DELETECompanyRowById)
	handler := c.Handler(mux)
	log.Fatal(http.ListenAndServe(":8000", handler))
}

func GETAllCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	rows, err := db.Query(`
            SELECT
			id, account, sys, username, pword, description, 
			address, "grouping", notes, aka, account_status
            FROM "Companies".companies
			ORDER BY id `)

	if err != nil {
		panic(err)
	}

	var company []Company

	for rows.Next() {
		var client Company
		rows.Scan(&client.Id, &client.Account, &client.Sys, &client.Username,
			&client.Pword, &client.Description, &client.Address,
			&client.Grouping, &client.Notes, &client.Aka, &client.Account_status)
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

	aka := r.URL.Query().Get("aka")
	println(aka)
	rows, err := db.Query(`SELECT 
		id, account, sys, username, pword, description, 
		address, "grouping", notes, aka, account_status
		FROM "Companies".companies
		WHERE aka in('` + aka + `')`)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var company []Company

	for rows.Next() {
		var client Company
		rows.Scan(&client.Id, &client.Account, &client.Sys, &client.Username,
			&client.Pword, &client.Description, &client.Address,
			&client.Grouping, &client.Notes, &client.Aka, &client.Account_status)

		if strings.Contains(strings.ToLower(client.Aka), strings.ToLower(aka)) {
			company = append(company, client)
		}
	}

	companyBytes, _ := json.MarshalIndent(company, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(companyBytes)

}

func UPDATEfield(w http.ResponseWriter, r *http.Request) {
	// update a field in the database
	w.Header().Set("Access-Control-Allow-Origin", "*")

	id, error := strconv.Atoi(r.URL.Query().Get("id"))
	value := r.URL.Query().Get("value")

	fmt.Println(id)
	fmt.Println(value)

	_, err := db.Exec(`
		UPDATE "Companies".companies	
		SET account=$2
		WHERE id=$1
		`, id, value)

	if err != nil || error != nil {
		panic(err)
	}

	fmt.Println("Successfully updated!")

}

func DELETECompanyRowById(w http.ResponseWriter, r *http.Request) {
	// update a field in the database
	w.Header().Set("Access-Control-Allow-Origin", "*")
	id, error := strconv.Atoi(r.URL.Query().Get("id"))

	fmt.Println(id)
	_, err := db.Exec(`
		DELETE FROM "Companies".companies	
		WHERE id = $1
		`, id)

	if err != nil || error != nil {
		panic(err)
	}

	fmt.Println("Successfully Deleted!")

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

func handleArgs() {
	flag.Parse()
	args := flag.Args()

	initDb()

	if len(args) >= 1 {
		switch args[0] {
		case "seed":
			seeds.Execute(db, args[1:]...)
			os.Exit(0)
		}
	}
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
	name, ok := "postgres", true
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
