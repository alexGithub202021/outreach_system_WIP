package lib

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// LoadProspectsFromCSV reads a semicolon-delimited CSV file whose path is
// taken from the CSV_PATH env var (default: ./prospects.csv), then inserts
// each row into the SQLite prospects table.
//
// Deduplication is done by email: a row is skipped if the email address
// already exists in the table.
//
// Expected CSV columns (header row required):
//
//	Gender ; Name ; Email ; Tmz ; Sending_date ; Sending_time ; Status
//
// Sending_time may be stored as `"23.59"` — dots are replaced with colons
// and the value is normalised to HH:MM:SS.
// Status "to_send" maps to status_p = 0.
func LoadProspectsFromCSV() error {

	// --- CSV path -----------------------------------------------------------
	csvPath := os.Getenv("CSV_PATH")
	if csvPath == "" {
		csvPath = "./prospects.csv"
	}

	f, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("LoadProspectsFromCSV: open %q: %w", csvPath, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.LazyQuotes = true // tolerates the triple-quoted time fields
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("LoadProspectsFromCSV: parse CSV: %w", err)
	}

	if len(records) < 2 {
		slog.Info("LoadProspectsFromCSV: CSV has no data rows, nothing to do")
		return nil
	}

	slog.Info("LoadProspectsFromCSV: CSV parsed",
		"path", csvPath,
		"data_rows", len(records)-1,
	)

	// --- SQLite connection --------------------------------------------------
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./SQLite/data/mydata.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("LoadProspectsFromCSV: open db %q: %w", dbPath, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("LoadProspectsFromCSV: ping db: %w", err)
	}
	slog.Info("LoadProspectsFromCSV: SQLite connected", "path", dbPath)

	// Build column-index map from the header row so the code is resilient to
	// column reordering.
	header := records[0]
	colIdx := make(map[string]int, len(header))
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	required := []string{"gender", "name", "email", "tmz", "sending_date", "sending_time", "status", "company"}
	for _, col := range required {
		if _, ok := colIdx[col]; !ok {
			return fmt.Errorf("LoadProspectsFromCSV: missing column %q in CSV header", col)
		}
	}

	inserted, skipped := 0, 0

	for rowNum, row := range records[1:] {
		// Pad short rows so index access is safe.
		for len(row) <= colIdx["sending_time"] {
			row = append(row, "")
		}

		gender := clean(row[colIdx["gender"]])
		name := clean(row[colIdx["name"]])
		email := clean(row[colIdx["email"]])
		tmz := clean(row[colIdx["tmz"]])
		sendingDate := clean(row[colIdx["sending_date"]])
		sendingTime := normaliseTime(clean(row[colIdx["sending_time"]]))
		statusRaw := clean(row[colIdx["status"]])
		company := clean(row[colIdx["company"]])

		// Skip rows with no email (cannot deduplicate or contact).
		if email == "" {
			slog.Warn("LoadProspectsFromCSV: skipping row with empty email", "row", rowNum+2)
			skipped++
			continue
		}

		// Convert status string to integer flag.
		statusP := 0
		if statusRaw != "to_send" {
			statusP = 1
		}

		// Deduplication check: skip if the email is already in the table.
		var existing int
		err := db.QueryRow(`SELECT COUNT(*) FROM prospects WHERE email = ? AND sending_date = ? AND sending_time = ?`, email, sendingDate, sendingTime).Scan(&existing)
		if err != nil {
			return fmt.Errorf("LoadProspectsFromCSV: dedup check row %d: %w", rowNum+2, err)
		}
		if existing > 0 {
			slog.Info("LoadProspectsFromCSV: duplicate skipped", "email", email, "sending_date", sendingDate, "sending_time", sendingTime)
			skipped++
			continue
		}

		_, err = db.Exec(
			`INSERT INTO prospects (name_p, email, gender, tmz, sending_date, sending_time, status_p, company)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			name, email, gender, tmz, sendingDate, sendingTime, statusP, company,
		)
		if err != nil {
			return fmt.Errorf("LoadProspectsFromCSV: insert row %d (%s): %w", rowNum+2, email, err)
		}

		slog.Info("LoadProspectsFromCSV: inserted", "name", name, "email", email)
		inserted++
	}

	slog.Info("LoadProspectsFromCSV: done", "inserted", inserted, "skipped", skipped)
	return nil
}

// clean strips surrounding whitespace and embedded newlines from a CSV field.
func clean(s string) string {
	// Remove embedded CR/LF that some spreadsheet exports include inside fields.
	s = strings.ReplaceAll(s, "\r\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	// Strip surrounding double-quotes left by triple-quoting artefacts.
	s = strings.Trim(s, "\"")
	return strings.TrimSpace(s)
}

// normaliseTime converts a time value that may arrive as "23.59" or "23:59"
// to the canonical HH:MM:SS format expected by the prospects query.
func normaliseTime(s string) string {
	// Replace dot-separators with colons (e.g. "23.59" → "23:59").
	s = strings.ReplaceAll(s, ".", ":")
	// Ensure HH:MM:SS — append ":00" if only HH:MM is present.
	if len(s) == 5 { // "HH:MM"
		s += ":00"
	}
	return s
}
