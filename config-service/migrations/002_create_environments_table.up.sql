CREATE TABLE IF NOT EXISTS environments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 50 CHECK (priority >= 0 AND priority <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_environments_name ON environments(name);
CREATE INDEX idx_environments_slug ON environments(slug);
CREATE INDEX idx_environments_active ON environments(active);
CREATE INDEX idx_environments_priority ON environments(priority);
CREATE INDEX idx_environments_created_at ON environments(created_at);

-- Create trigger for updated_at
CREATE TRIGGER update_environments_updated_at
    BEFORE UPDATE ON environments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default environments
INSERT INTO environments (name, slug, description, priority) VALUES
    ('Development', 'dev', 'Development environment for testing new features', 10),
    ('Staging', 'staging', 'Staging environment for pre-production testing', 50),
    ('Production', 'prod', 'Production environment for live applications', 90)
ON CONFLICT (slug) DO NOTHING;
