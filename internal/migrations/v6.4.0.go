package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V6_4_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	if _, err := db.Exec(`INSERT INTO settings (key, value) VALUES
		('upload.r2.account_id', '""'),
		('upload.r2.access_key_id', '""'),
		('upload.r2.secret_access_key', '""'),
		('upload.r2.bucket', '""'),
		('upload.r2.bucket_path', '"/"'),
		('upload.r2.public_url', '""'),
		('upload.r2.expiry', '"167h"')
		ON CONFLICT (key) DO NOTHING`); err != nil {
		return err
	}

	return nil
}
