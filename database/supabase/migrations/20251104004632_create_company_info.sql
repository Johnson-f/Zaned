-- Create company_info table
CREATE TABLE IF NOT EXISTS company_info (
    symbol VARCHAR(20) PRIMARY KEY NOT NULL,
    name VARCHAR(255) NOT NULL,
    price VARCHAR(50),
    after_hours_price VARCHAR(50),
    change VARCHAR(50),
    percent_change VARCHAR(50),
    open VARCHAR(50),
    high VARCHAR(50),
    low VARCHAR(50),
    year_high VARCHAR(50),
    year_low VARCHAR(50),
    volume BIGINT,
    avg_volume BIGINT,
    market_cap VARCHAR(50),
    beta VARCHAR(50),
    pe VARCHAR(50),
    earnings_date VARCHAR(100),
    sector VARCHAR(255),
    industry VARCHAR(255),
    about TEXT,
    employees VARCHAR(50),
    five_days_return VARCHAR(50),
    one_month_return VARCHAR(50),
    three_month_return VARCHAR(50),
    six_month_return VARCHAR(50),
    ytd_return VARCHAR(50),
    year_return VARCHAR(50),
    three_year_return VARCHAR(50),
    five_year_return VARCHAR(50),
    ten_year_return VARCHAR(50),
    max_return VARCHAR(50),
    logo TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_company_info_name ON company_info(name);
CREATE INDEX IF NOT EXISTS idx_company_info_sector ON company_info(sector);
CREATE INDEX IF NOT EXISTS idx_company_info_industry ON company_info(industry);
CREATE INDEX IF NOT EXISTS idx_company_info_deleted_at ON company_info(deleted_at);

-- Enable Row Level Security (RLS) on the table
ALTER TABLE company_info ENABLE ROW LEVEL SECURITY;

-- RLS Policies for company_info table
-- Only allow SELECT operations (read-only for all users)
-- INSERT, UPDATE, and DELETE are automatically blocked by RLS (no policies = blocked)

-- Drop existing policies if they exist (for idempotency)
DROP POLICY IF EXISTS "Allow select on company info" ON company_info;

-- Policy: All users can view company info (read-only access)
-- No policies for INSERT/UPDATE/DELETE means they are automatically blocked
CREATE POLICY "Allow select on company info"
    ON company_info
    FOR SELECT
    USING (true);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_company_info_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_company_info_updated_at
    BEFORE UPDATE ON company_info
    FOR EACH ROW
    EXECUTE FUNCTION update_company_info_updated_at();

-- Enable real-time subscription for company_info table
-- ALTER PUBLICATION supabase_realtime ADD TABLE company_info;


-- Policy for the screener_result table
-- Enable Row Level Security (RLS) on the table
ALTER TABLE screener_result ENABLE ROW LEVEL SECURITY;

-- RLS Policies for screener_result table
-- Only allow SELECT operations (read-only for all users)
-- INSERT, UPDATE, and DELETE are automatically blocked by RLS (no policies = blocked)

-- Drop existing policies if they exist (for idempotency)
DROP POLICY IF EXISTS "Allow select on screener result" ON screener_result;

-- Policy: All users can view screener result (read-only access)
-- No policies for INSERT/UPDATE/DELETE means they are automatically blocked
CREATE POLICY "Allow select on screener result"
    ON screener_result
    FOR SELECT
    USING (true);