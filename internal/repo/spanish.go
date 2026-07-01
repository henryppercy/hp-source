package repo

import "fmt"

// DreamingSpanishDay is a single day's input time pulled from Dreaming Spanish.
type DreamingSpanishDay struct {
	Date    string
	Seconds int
}

// ReplaceDreamingSpanish swaps every synced Dreaming Spanish row for the given
// days in one transaction. Delete-and-replace keeps sync idempotent and matches
// DS even when it corrects past days; manually logged rows are left untouched.
func (r *Repo) ReplaceDreamingSpanish(days []DreamingSpanishDay) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM spanish_log WHERE source = 'dreaming_spanish'`); err != nil {
		return fmt.Errorf("failed to clear dreaming spanish rows: %w", err)
	}

	for _, day := range days {
		_, err := tx.Exec(
			`INSERT INTO spanish_log (date, seconds, activity, source)
             VALUES (?, ?, 'ci', 'dreaming_spanish')`,
			day.Date, day.Seconds,
		)
		if err != nil {
			return fmt.Errorf("failed to insert %s: %w", day.Date, err)
		}
	}

	return tx.Commit()
}
