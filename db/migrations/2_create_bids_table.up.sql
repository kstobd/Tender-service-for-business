
CREATE TABLE bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('CREATED', 'PUBLISHED', 'CANCELED')), 
    decision VARCHAR(20) CHECK (decision IN ('Approved', 'Rejected')),
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    author_type VARCHAR(20) NOT NULL CHECK (author_type IN ('Organization', 'User')),
    author_id UUID NOT NULL,  -- Идентификатор автора (организация или пользователь)
    version INTEGER NOT NULL DEFAULT 1,  -- Номер версии
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
