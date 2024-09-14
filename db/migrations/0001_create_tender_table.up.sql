CREATE TABLE tender (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    service_type VARCHAR(50) NOT NULL,
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    creator_id UUID REFERENCES employee(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('CREATED', 'PUBLISHED', 'CLOSED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
