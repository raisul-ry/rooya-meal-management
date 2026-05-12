package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SendMealCount posts tomorrow's meal count to a Teams incoming webhook.
func SendMealCount(webhookURL, date string, count int) error {
	t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	label := t.Format("Monday, January 02")

	msg := map[string]string{
		"text": fmt.Sprintf("**Meal count for %s: %d meal(s)**", label, count),
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams webhook returned %d", resp.StatusCode)
	}
	return nil
}
