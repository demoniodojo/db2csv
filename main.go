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
	"log"
)


func main () {

	// define the flags
	userFlag 	 := flag.String("user", "", "Database username")
	passwordFlag := flag.String("password", "", "password for the username")
	hostFlag	 := flag.String("host", "127.0.0.1", "Hostname of the database")
	portFlag	 := flag.Int("port", 3306, "Host port for connection")
	filterFlag	 := flag.String("filter", "", "parameter for the WHERE clause")
	noHeaderFlag := flag.Bool("no-header", false, "Skip writing the column names as the first row")

	// improve the -help on usage
	flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
			fmt.Println("Example: db2csv -user root -password secret mydb.mytable output.csv")
			fmt.Println("\nIf output.csv is omitted, it defaults to tableName.csv")
			fmt.Println("\nAvailable Flags:")
			flag.PrintDefaults()
	}

	// process the flags
	flag.Parse()

	// checking for all arguments
	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	// Parsing
	databaseName, tableName, err := parseSource(flag.Arg(0))
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}

	fileName := ""
	if len(flag.Args()) > 1 {
		fileName = strings.TrimSpace(flag.Arg(1))
	} else {
		fileName = tableName + ".csv"
	}

	// Connecting
	db, err := connectToDB(*userFlag, *passwordFlag, databaseName, *hostFlag, *portFlag)
	if err != nil {
		fmt.Println("Database Connection Error:", err)
		return
	}

	defer db.Close()


	fmt.Printf("Working with\n Database: %s\n Table: %s\n File: %s\n", databaseName, tableName, fileName)

	err = exportToCSV(db, tableName, fileName, *filterFlag, *noHeaderFlag)
	if err != nil {
		fmt.Println(err)
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
func connectToDB(user, password, databaseName, host string, port int) (*sql.DB, error) {
	
	// Sanitize credential
	safeUser 		:= url.QueryEscape(user)
	safePassword	:= url.QueryEscape(password)
	safeHost		:= url.QueryEscape(host)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", safeUser, safePassword, safeHost, port, databaseName)

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

func exportToCSV (db *sql.DB, tableName, fileName, filter string, noHeader bool) error {
	
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Error, could not create output file [%s]: %w", fileName, err)
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
		return fmt.Errorf("Error with the query [%s]: %w", query, err)
	}
	defer rows.Close()

	// discover header, column field name
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("Error, failed to get column names: %w", err)
	}

	if !noHeader {
		// write headers if not asked otherwise
		if err = writer.Write(columns); err != nil {
			return err
		}
	}

	// traverse the recordset
	values 	  := make([]interface{}, len(columns)) // build an array for the data

	valuePtrs := make([]interface{}, len(columns)) // build an array for the addresses (pointers)

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

	// check if there were errors during the export
	if err = rows.Err(); err != nil {
        return fmt.Errorf("Error during row iteration: %w", err)
    }

	writer.Flush()
    if err := writer.Error(); err != nil {
        return fmt.Errorf("Error, failed to finalize CSV file: %w", err)
    }
	
	return nil
}