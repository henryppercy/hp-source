package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

const dreamingSpanishURL = "https://app.dreaming.com/.netlify/functions/dayWatchedTime?language=es"

func newSpanishCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spanish",
		Short: "Track your Spanish comprehensible input",
	}
	cmd.AddCommand(newSpanishSyncCmd(a))
	return cmd
}

func newSpanishSyncCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Pull logged time from Dreaming Spanish into the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSpanishSync(a.repo)
		},
	}
}

func runSpanishSync(r *repo.Repo) error {
	token := os.Getenv("DS_BEARER_TOKEN")
	if token == "" {
		return fmt.Errorf("DS_BEARER_TOKEN is not set")
	}

	days, err := fetchDreamingSpanish(token)
	if err != nil {
		return err
	}

	if err := r.ReplaceDreamingSpanish(days); err != nil {
		return err
	}

	fmt.Printf("synced %d days from Dreaming Spanish\n", len(days))
	return nil
}

// fetchDreamingSpanish returns every day with logged time, dropping empty days
// since zero rows are derivable from the calendar at render time.
func fetchDreamingSpanish(token string) ([]repo.DreamingSpanishDay, error) {
	req, err := http.NewRequest(http.MethodGet, dreamingSpanishURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Dreaming Spanish: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dreaming spanish returned %s", res.Status)
	}

	var payload []struct {
		Date        string  `json:"date"`
		TimeSeconds float64 `json:"timeSeconds"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var days []repo.DreamingSpanishDay
	for _, entry := range payload {
		if entry.TimeSeconds <= 0 {
			continue
		}
		days = append(days, repo.DreamingSpanishDay{Date: entry.Date, Seconds: int(math.Round(entry.TimeSeconds))})
	}
	return days, nil
}
