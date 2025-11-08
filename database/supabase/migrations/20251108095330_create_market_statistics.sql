-- Create market_statistics table
CREATE TABLE IF NOT EXISTS market_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL UNIQUE,
    stocks_up INTEGER NOT NULL DEFAULT 0,
    stocks_down INTEGER NOT NULL DEFAULT 0,
    stocks_unchanged INTEGER NOT NULL DEFAULT 0,
    total_stocks INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_market_statistics_date ON market_statistics(date);
CREATE INDEX IF NOT EXISTS idx_market_statistics_deleted_at ON market_statistics(deleted_at);

-- Enable Row Level Security (RLS) on the table
ALTER TABLE market_statistics ENABLE ROW LEVEL SECURITY;

-- RLS Policies for market_statistics table
-- Only allow SELECT operations (read-only for all users)
-- INSERT, UPDATE, and DELETE are automatically blocked by RLS (no policies = blocked)

-- Drop existing policies if they exist (for idempotency)
DROP POLICY IF EXISTS "Allow select on market statistics" ON market_statistics;

-- Policy: All users can view market statistics (read-only access)
-- No policies for INSERT/UPDATE/DELETE means they are automatically blocked
CREATE POLICY "Allow select on market statistics"
    ON market_statistics
    FOR SELECT
    USING (true);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_market_statistics_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_market_statistics_updated_at
    BEFORE UPDATE ON market_statistics
    FOR EACH ROW
    EXECUTE FUNCTION update_market_statistics_updated_at();

