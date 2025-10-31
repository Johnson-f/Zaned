package supabase

import (
	"fmt"
	"os"

	"github.com/supabase-community/supabase-go"
)

var Client *supabase.Client

// InitClient initializes the Supabase client with URL and anon key from environment variables
func InitClient() error {
	url := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	if url == "" || anonKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	client, err := supabase.NewClient(url, anonKey, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	Client = client
	return nil
}

