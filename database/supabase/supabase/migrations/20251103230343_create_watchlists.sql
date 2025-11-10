-- Enable Row Level Security (RLS) on both tables
ALTER TABLE watchlists ENABLE ROW LEVEL SECURITY;
ALTER TABLE watchlist_items ENABLE ROW LEVEL SECURITY;

-- RLS Policies for watchlists table

-- Drop existing policies if they exist (for idempotency)
DROP POLICY IF EXISTS "Users can view their own watchlists" ON watchlists;
DROP POLICY IF EXISTS "Users can insert their own watchlists" ON watchlists;
DROP POLICY IF EXISTS "Users can update their own watchlists" ON watchlists;
DROP POLICY IF EXISTS "Users can delete their own watchlists" ON watchlists;

-- Policy: Users can view their own watchlists
CREATE POLICY "Users can view their own watchlists"
    ON watchlists
    FOR SELECT
    USING (auth.uid() = user_id);

-- Policy: Users can insert their own watchlists
CREATE POLICY "Users can insert their own watchlists"
    ON watchlists
    FOR INSERT
    WITH CHECK (auth.uid() = user_id);

-- Policy: Users can update their own watchlists
CREATE POLICY "Users can update their own watchlists"
    ON watchlists
    FOR UPDATE
    USING (auth.uid() = user_id)
    WITH CHECK (auth.uid() = user_id);

-- Policy: Users can delete their own watchlists
CREATE POLICY "Users can delete their own watchlists"
    ON watchlists
    FOR DELETE
    USING (auth.uid() = user_id);

-- RLS Policies for watchlist_items table

-- Drop existing policies if they exist (for idempotency)
DROP POLICY IF EXISTS "Users can view items from their own watchlists" ON watchlist_items;
DROP POLICY IF EXISTS "Users can insert items into their own watchlists" ON watchlist_items;
DROP POLICY IF EXISTS "Users can update items in their own watchlists" ON watchlist_items;
DROP POLICY IF EXISTS "Users can delete items from their own watchlists" ON watchlist_items;

-- Policy: Users can view items from their own watchlists
CREATE POLICY "Users can view items from their own watchlists"
    ON watchlist_items
    FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM watchlists
            WHERE watchlists.id = watchlist_items.watchlist_id
            AND watchlists.user_id = auth.uid()
        )
    );

-- Policy: Users can insert items into their own watchlists
CREATE POLICY "Users can insert items into their own watchlists"
    ON watchlist_items
    FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM watchlists
            WHERE watchlists.id = watchlist_items.watchlist_id
            AND watchlists.user_id = auth.uid()
        )
    );

-- Policy: Users can update items in their own watchlists
CREATE POLICY "Users can update items in their own watchlists"
    ON watchlist_items
    FOR UPDATE
    USING (
        EXISTS (
            SELECT 1 FROM watchlists
            WHERE watchlists.id = watchlist_items.watchlist_id
            AND watchlists.user_id = auth.uid()
        )
    )
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM watchlists
            WHERE watchlists.id = watchlist_items.watchlist_id
            AND watchlists.user_id = auth.uid()
        )
    );

-- Policy: Users can delete items from their own watchlists
CREATE POLICY "Users can delete items from their own watchlists"
    ON watchlist_items
    FOR DELETE
    USING (
        EXISTS (
            SELECT 1 FROM watchlists
            WHERE watchlists.id = watchlist_items.watchlist_id
            AND watchlists.user_id = auth.uid()
        )
    );

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_watchlists_updated_at
    BEFORE UPDATE ON watchlists
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_watchlist_items_updated_at
    BEFORE UPDATE ON watchlist_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable real-time subscription for watchlists table
ALTER PUBLICATION supabase_realtime ADD TABLE watchlists;

-- Enable real-time subscription for watchlist_items table
ALTER PUBLICATION supabase_realtime ADD TABLE watchlist_items;

