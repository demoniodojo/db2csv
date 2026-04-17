package main

import (
	"database/sql"
	"encoding/csv"
	"fmt" 		// basic
	"os" 		// for arguments in the command
	"strings"	// for strings management
	"flag"		// for flags in the command
	_ "github.com/go-sql-driver/mysql" // for mysql connection, the _ is for it to call init
	"net/url" // for dealing with weird chars in passwords and similar
)


func main () {

	// define the flags
	userFlag 	 := flag.String("user", "", "Database username")
	passwordFlag := flag.String("password", "", "password for the username")
	hostFlag	 := flag.String("host", "127.0.0.1", "Hostname of the database")
	portFlag	 := flag.String("port", "3306", "Host port for connection")
	filterFlag	 := flag.String("filter", "", "parameter for the WHERE clause")

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
	db, err := connectToDB(*userFlag, *passwordFlag, databaseName, *hostFlag, *portFlag)
	if err != nil {
		fmt.Println("Database Connection Error:", err)
		return
	}

	defer db.Close()


	fmt.Printf("Working with\n Database: %s\n Table: %s\n File: %s\n", databaseName, tableName, fileName)

	err = exportToCSV(db, tableName, fileName, *filterFlag)
	if err != nil {
		fmt.Println("Error creating CSV:", err)
		return
	}
	fmt.Println("File created successfully!")
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
func connectToDB(user, password, databaseName, host, port string) (*sql.DB, error) {
	
	// Sanitize credential
	safeUser 		:= url.QueryEscape(user)
	safePassword	:= url.QueryEscape(password)
	safeHost		:= url.QueryEscape(host)
	safePort		:= url.QueryEscape(port)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", safeUser, safePassword, safeHost, safePort, databaseName)

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

func exportToCSV (db *sql.DB, tableName, fileName, filter string) error {
	
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	if filter != "" {
		query = fmt.Sprintf("%s WHERE %s", query, filter)
	}

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	} 

	// write headers
	err = writer.Write(columns)
	if err != nil {
		return err
	}

	// traverse the recordset
	values 	  := make([]interface{}, len(columns))

	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		// scan db row in values slice
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return err
		}

		// convert the values into strings for the CSV
		rowRecord := make([]string, len(columns))
		for i, val := range values {
		    if val == nil {
		        rowRecord[i] = ""
		    } else {
		        // Check if the value is actually a slice of bytes
		        if b, ok := val.([]byte); ok {
		            rowRecord[i] = string(b) // Convert []byte to string
		        } else {
		            rowRecord[i] = fmt.Sprintf("%v", val) // Handle ints, floats, etc.
		        }
		    }
		}

		writer.Write(rowRecord)
	}

	return nil
}