-- Users table with Argon2id migration support
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified_at TIMESTAMP,
    password VARCHAR(255) NOT NULL, -- Argon2id hash
    password_reset_required BOOLEAN DEFAULT false,
    reset_token_hash VARCHAR(64), -- SHA256 hash of reset token
    reset_token_expires_at TIMESTAMP,
    remember_token VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id BIGSERIAL PRIMARY KEY,
    status SMALLINT NOT NULL DEFAULT 1,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255),
    client VARCHAR(255),
    "order" INTEGER NOT NULL DEFAULT 0,
    highlighted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Project Images
CREATE TABLE project_images (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL, -- /storage/img/filename.jpg
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    "order" INTEGER NOT NULL DEFAULT 0,
    blur_hash TEXT, -- data URL format
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_project_images_project_id ON project_images(project_id);

-- Testimonials
CREATE TABLE testimonials (
    id BIGSERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    profession VARCHAR(255) NOT NULL,
    testimonial TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ready',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Static Texts
CREATE TABLE static_texts (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    label VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Configurations
CREATE TABLE configurations (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Visitor Messages
CREATE TABLE visitor_messages (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    description TEXT NOT NULL,
    seen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed configurations
INSERT INTO configurations (key, value) VALUES ('config.site.live', 'false');

-- Seed static texts (known keys)
INSERT INTO static_texts (key, label, content) VALUES
    ('hero.title', 'Hero Title', 'Welcome to Elite Constructions'),
    ('hero.subtitle', 'Hero Subtitle', 'Building Excellence Since...'),
    ('projects.read_more', 'Read More', 'Read More'),
    ('company.name', 'Company Name', 'Elite Constructions'),
    ('company.description', 'Company Description', '...'),
    ('footer.copyright', 'Footer Copyright', '© 2024 Elite Constructions');
