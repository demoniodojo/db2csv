# db2csv

A lightweight, high-performance CLI tool written in Go to export MySQL/MariaDB tables directly to CSV files. Designed for speed, portability, and ease of use in Linux/WSL environments.

## Features

- **Zero Dependencies:** Compiles to a single static binary. No need for Go, Python, or PHP runtimes on the target server.
- **Efficient Memory Usage:** Streams data row-by-row, allowing for the export of massive datasets without exhausting RAM.
- **Smart Type Handling:** Automatically handles MySQL types, including converting `[]byte` (blobs/text) to strings and managing `NULL` values.
- **Flexible Filtering:** Supports raw SQL `WHERE` clauses via a `--filter` flag.
- **Secure:** Supports environment variables and URL-encoded credentials for safe database connections.

## Installation

### From Source
1. Ensure you have [Go](https://go.dev/doc/install) installed.
2. Clone the repository:
   ```bash
   git clone https://github.com/demoniodojo/db2csv.git]
   cd db2csv
   ```
3. Build the binary:
   ```bash
   go build -o db2csv main.go
   ```
4. (Optional) Move to your path:
   ```bash
   sudo mv db2csv /usr/local/bin/
   ```

## Usage

Basic command structure:
```bash
db2csv [flags] <database.table> <output.csv>
```

### Examples

**Standard Export:**
```bash
db2csv --user root --password mysecret my_database.users users_backup.csv
```

**Using a Filter:**
```bash
db2csv --user admin --password pass --filter "status='active' AND created_at > '2023-01-01'" web_db.orders active_orders.csv
```

**Remote Connection:**
```bash
db2csv --host 192.168.1.50 --port 3307 --user remote_user mydb.products products.csv
```

### Flags
- `--user`: Database username (required).
- `--password`: Database password (required).
- `--host`: Database hostname (default: `127.0.0.1`).
- `--port`: Database port (default: `3306`).
- `--filter`: Optional SQL snippet for the `WHERE` clause.

## Technical Details

- **Language:** Go (Golang)
- **Database Driver:** `github.com/go-sql-driver/mysql`
- **Architecture:** - Uses `sql.RawBytes` and `interface{}` slices for dynamic row scanning, enabling compatibility with any table schema without pre-definition.
    - Implements `Type Assertion` to safely convert raw database bytes into UTF-8 strings.
    - Utilizes `defer` stacks for guaranteed resource cleanup (database handles and file buffers).

## License
MIT