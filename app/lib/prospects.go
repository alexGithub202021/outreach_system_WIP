package lib

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// Prospect holds a single row from the prospects table.
type Prospect struct {
	ID          int
	Name        string // name_p column
	Email       string // email column
	Gender      string // gender column: "W" = female, anything else = male
	SendingDate string // stored as "YYYY-MM-DD"
	SendingTime string // stored as "HH:MM:SS"
	Company     string // company column
}

// ProcessProspects opens the SQLite database, reads all prospects whose
// status_p = 0, sending_date <= today, and sending_time <= current time,
// and logs each matching prospect with slog.
//
// The DB path is read from the DB_PATH environment variable; if unset it
// falls back to the local dev path used in main.go.
func ProcessProspects() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./SQLite/data/mydata.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("ProcessProspects -> open db %q: %w", dbPath, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ProcessProspects -> ping db %q: %w", dbPath, err)
	}
	slog.Info("ProcessProspects -> SQLite connected", "path", dbPath)

	// Load the timezone from the TZ env var (set in docker-compose).
	// This ensures time.Now() reflects the timezone where the data was recorded,
	// not the container's default UTC clock.
	tzName := os.Getenv("TZ")
	if tzName == "" {
		tzName = "UTC"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		slog.Warn("ProcessProspects -> unknown TZ, falling back to UTC", "TZ", tzName, "err", err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := now.Format("2006-01-02")     // "YYYY-MM-DD"
	currentTime := now.Format("15:04:05") // "HH:MM:SS"
	slog.Info("ProcessProspects -> using timezone", "TZ", loc.String(), "today", today, "currentTime", currentTime)

	// Select prospects that are pending (status_p = 0),
	// whose sending_date is earlier,
	// OR sending_date is equal AND whose sending_time is now or earlier.
	query := `
		SELECT id, name_p, email, gender, sending_date, sending_time, company
		FROM   prospects
		WHERE  status_p = 0
		  AND  (sending_date < ? OR (sending_date = ? AND sending_time <= ?))`

	slog.Info("ProcessProspects -> filtering prospects",
		"filtering query", query,
		"param_sending_date", today,
		"param_sending_date", today,
		"param_sending_time", currentTime,
	)

	rows, err := db.Query(query, today, today, currentTime)
	if err != nil {
		return fmt.Errorf("ProcessProspects -> query: %w", err)
	}
	defer rows.Close()

	// list of prospects whose get email sent
	var prospectIds []int

	count := 0
	for rows.Next() {
		var p Prospect

		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Gender, &p.SendingDate, &p.SendingTime, &p.Company); err != nil {
			return fmt.Errorf("ProcessProspects -> scan row: %w", err)
		}

		slog.Info("ProcessProspects -> sending email to: ",
			"name", p.Name,
			"email", p.Email,
			"company", p.Company,
			"date to send", p.SendingDate,
			"time to send", p.SendingTime,
			"id", p.ID,
		)

		// todo -> add sleep, tamper emails sending -> avoid being flagged by spam filters

		sleep(30)

		// Send the prospecting email.
		if !CallResendApi(p.Email, CapitalizeFirst(p.Name), CapitalizeFirst(p.Company)) {
			slog.Error("ProcessProspects -> email sending failed",
				"id", p.ID,
				"email", p.Email,
				"err", err,
			)
			// Continue with next prospect rather than aborting the whole run.
			continue
		}

		// update list prospects with msg already sent
		prospectIds = append(prospectIds, p.ID)

		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("ProcessProspects -> rows iteration: %w", err)
	}

	// Mark as sent so the prospect is never emailed again.
	updateProspectsStatus(db, prospectIds)

	slog.Info("ProcessProspects -> done", "number of emails sent:", count)
	return nil
}

func updateProspectsStatus(db *sql.DB, prospects []int) {

	for _, prospect := range prospects {

		query := `UPDATE prospects SET status_p = 1 WHERE id = ?`

		slog.Info("ProcessProspects -> update query",
			"query", query,
			"prospect ID", prospect,
		)

		_, err := db.Query(query, prospect)
		if err != nil {
			slog.Error("ProcessProspects -> failed to update status_p for:",
				"prospect id", prospect,
				"err", err,
			)
		}

		slog.Info("updateProspectsStatus -> successful", "prospect ID: ", prospect)
	}

}

func sleep(dur int) {
	duration := time.Duration(dur)

	fmt.Println("Pause...")
	time.Sleep(duration * time.Second) // Pause for 5 seconds
	fmt.Println("Resuming...")
}
