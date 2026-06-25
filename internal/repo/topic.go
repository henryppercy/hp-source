package repo

import "fmt"

type Topic struct {
	ID   int
	Name string
}

func (r *Repo) ListTopics() ([]Topic, error) {
	rows, err := r.db.Query("SELECT id, name FROM topic ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var t Topic
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("failed to scan topic: %w", err)
		}
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

func linkPostTopic(tx TX, postID, topicID int) error {
	_, err := tx.Exec(
		"INSERT INTO post_topic (post_id, topic_id) VALUES (?, ?)",
		postID, topicID,
	)
	return err
}
