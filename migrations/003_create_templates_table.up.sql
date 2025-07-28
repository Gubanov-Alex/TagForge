-- Create ENUM for config format
CREATE TYPE config_format AS ENUM ('json', 'yaml', 'toml', 'env');

CREATE TABLE IF NOT EXISTS templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    format config_format NOT NULL DEFAULT 'json',
    content TEXT NOT NULL,
    schema JSONB DEFAULT '{}',
    default_values JSONB DEFAULT '{}',
    version VARCHAR(50) NOT NULL,
    environment_id BIGINT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(100) NOT NULL,
    updated_by VARCHAR(100) NOT NULL,

    -- Unique constraint for name + environment
    UNIQUE(name, environment_id)
);

-- Create template_tags junction table for many-to-many relationship
CREATE TABLE IF NOT EXISTS template_tags (
    template_id BIGINT NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (template_id, tag_id)
);

-- Create indexes for templates
CREATE INDEX idx_templates_name ON templates(name);
CREATE INDEX idx_templates_environment_id ON templates(environment_id);
CREATE INDEX idx_templates_format ON templates(format);
CREATE INDEX idx_templates_version ON templates(version);
CREATE INDEX idx_templates_active ON templates(active);
CREATE INDEX idx_templates_created_at ON templates(created_at);
CREATE INDEX idx_templates_created_by ON templates(created_by);

-- Create GIN indexes for JSONB columns
CREATE INDEX idx_templates_schema_gin ON templates USING GIN (schema);
CREATE INDEX idx_templates_default_values_gin ON templates USING GIN (default_values);

-- Create indexes for template_tags
CREATE INDEX idx_template_tags_template_id ON template_tags(template_id);
CREATE INDEX idx_template_tags_tag_id ON template_tags(tag_id);

-- Create trigger for updated_at
CREATE TRIGGER update_templates_updated_at
    BEFORE UPDATE ON templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO tags (name, description, color) VALUES
    ('database', 'Database configuration templates', '#3b82f6'),
    ('api', 'API service configuration templates', '#10b981'),
    ('monitoring', 'Monitoring and observability configs', '#f59e0b'),
    ('security', 'Security and authentication configs', '#ef4444')
ON CONFLICT (name) DO NOTHING;
