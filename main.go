package main

import (
	"database/sql"
	"fmt" 		// basic
	// "os" 		// for arguments in the command
	"strings"	// for strings management
	"flag"		// for flags in the command
	_ "github.com/go-sql-driver/mysql" // for mysql connection, the _ is for it to call init
	"net/url" // for dealing with : in passwords and similar
)


func main () {

	// define the flags
	rawUser 	 := flag.String("user", "", "Database username")
	rawPassword := flag.String("password", "", "password for the username")

	// process the flags
	flag.Parse()

	// sanitizing password
	user := url.QueryEscape(*rawUser)
	password := url.QueryEscape(*rawPassword)


	// checking for all arguments
	if len(flag.Args()) < 2 {
		fmt.Println("Error. Usage: <database.table> <destination_file>")
		return
	}

	// splitting the database.table 
	source := strings.Split(flag.Arg(0), ".")

	// validating that the source is correctly formed
	if len(source) != 2 {
		fmt.Println("Error. The source format must be database.table. Example: my_Database.my_Table")
		return
	}

	if source[0] == "" || source[1] == "" {
		fmt.Println("Error: both database and table must be provided.")
		return
	}

	databaseName := strings.TrimSpace(source[0])
	tableName := strings.TrimSpace(source[1])
	fileName := strings.TrimSpace(flag.Arg(1))

	fmt.Printf("Working with\n Database: %s\n Table: %s\n File: %s\n User: %s\n Password: %s\n", databaseName, tableName, fileName, user, password)

	// Building the connector
	host := "127.0.0.1"
	port := "3306"

	dsn	 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, databaseName)

	// initialize the pool object
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	// schedule the cleanup
	defer db.Close()

	// configure the pool
	db.SetMaxOpenConns(1)

	// test the connection
	err = db.Ping()
    if err != nil {
        fmt.Println("Connection failed:", err)
        return
    }

}