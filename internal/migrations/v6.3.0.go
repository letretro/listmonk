package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V6_3_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	if _, err := db.Exec(`INSERT INTO settings (key, value) VALUES ('zeptomail', '[{"name":"ZeptoMail","enabled":false,"api_key":"","from_email":"","from_name":"","track_opens":true,"track_clicks":true,"timeout":"30s","max_msg_retries":2}]')
		ON CONFLICT (key) DO NOTHING`); err != nil {
		return err
	}

	return nil
}
