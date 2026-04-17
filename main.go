package main

import (
	"database/sql"
	"fmt" 		// basic
	// "os" 		// for arguments in the command
	"strings"	// for strings management
	"flag"		// for flags in the command
	_ "github.com/go-sql-driver/mysql" // for mysql connection, the _ is for it to call init
	"net/url" // for dealing with weird chars in passwords and similar
)


func main () {

	// define the flags
	userFlag 	 := flag.String("user", "", "Database username")
	passwordFlag := flag.String("password", "", "password for the username")

	// process the flags
	flag.Parse()

	// checking for all arguments
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: db2csv [flags] <database.table> <output.csv>")
		return
	}

	// Parsing
	databaseName, tableName, err := parseSource(flag.Arg(0))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fileName 	 := strings.TrimSpace(flag.Arg(1))


	// Connecting
	db, err := connectToDB(*userFlag, *passwordFlag, databaseName)
	if err != nil {
		fmt.Println("Database Connection Error:", err)
		return
	}

	defer db.Close()


	fmt.Printf("Working with\n Database: %s\n Table: %s\n File: %s\n", databaseName, tableName, fileName)

}

// Function to parse the input
func parseSource(input string) (string, string, error) {
	source := strings.Split(input, ".")
	if len(source) != 2 {
		return "", "", fmt.Errorf("invalid source format: must be database.table (got: %s)", input)
	}

	databaseName := strings.TrimSpace(source[0])
	tableName	 := strings.TrimSpace(source[1])

	if databaseName == "" || tableName == "" {
		return "", "", fmt.Errorf("database and table names cannot be empty")
	}

	return databaseName, tableName, nil
}

// Function to connect to DB
func connectToDB(user, password, databaseName string) (*sql.DB, error) {
	host := "127.0.0.1"
	port := "3306"

	// Sanitize credential
	safeUser 		:= url.QueryEscape(user)
	safePassword	:= url.QueryEscape(password)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", safeUser, safePassword, host, port, databaseName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}