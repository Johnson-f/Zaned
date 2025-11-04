-- Create fundamental_data table
CREATE TABLE IF NOT EXISTS fundamental_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    statement_type VARCHAR(50) NOT NULL,
    frequency VARCHAR(20) NOT NULL,
    statement JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(symbol, statement_type, frequency)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_fundamental_symbol ON fundamental_data(symbol);
CREATE INDEX IF NOT EXISTS idx_fundamental_statement_type ON fundamental_data(statement_type);
CREATE INDEX IF NOT EXISTS idx_fundamental_frequency ON fundamental_data(frequency);
CREATE INDEX IF NOT EXISTS idx_fundamental_symbol_type_freq ON fundamental_data(symbol, statement_type, frequency);
CREATE INDEX IF NOT EXISTS idx_fundamental_deleted_at ON fundamental_data(deleted_at);

-- Create GIN index on JSONB column for efficient JSON queries
CREATE INDEX IF NOT EXISTS idx_fundamental_statement_gin ON fundamental_data USING GIN (statement);

-- Enable Row Level Security (RLS) on the table
ALTER TABLE fundamental_data ENABLE ROW LEVEL SECURITY;

-- RLS Policies for fundamental_data table
-- Only allow SELECT operations (read-only for all users)
-- INSERT, UPDATE, and DELETE are automatically blocked by RLS (no policies = blocked)

-- Drop existing policies if they exist (for idempotency)
DROP POLICY IF EXISTS "Allow select on fundamental data" ON fundamental_data;

-- Policy: All users can view fundamental data (read-only access)
-- No policies for INSERT/UPDATE/DELETE means they are automatically blocked
CREATE POLICY "Allow select on fundamental data"
    ON fundamental_data
    FOR SELECT
    USING (true);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_fundamental_data_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_fundamental_data_updated_at
    BEFORE UPDATE ON fundamental_data
    FOR EACH ROW
    EXECUTE FUNCTION update_fundamental_data_updated_at();

-- Enable real-time subscription for fundamental_data table
ALTER PUBLICATION supabase_realtime ADD TABLE fundamental_data;

