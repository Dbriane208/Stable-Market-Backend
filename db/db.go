package db

import (
	"errors"
	"os"

	supa "github.com/nedpals/supabase-go"
)

var Supabase *supa.Client

// DatabaseClient initializes the Supabase client
func DatabaseClient() error {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	if url == "" || key == "" {
		return errors.New("SUPABASE_URL and SUPABASE_KEY must be set")
	}

	Supabase = supa.CreateClient(url, key)
	return nil
}
