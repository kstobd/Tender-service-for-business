CREATE TABLE tender (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    service_type VARCHAR(50) NOT NULL CHECK (service_type IN ('Construction', 'Delivery', 'Manufacture')),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,  -- связь с organization
    creator_id UUID REFERENCES employee(id) ON DELETE CASCADE,  -- связь с employee
    responsible_id UUID REFERENCES organization_responsible(id) ON DELETE CASCADE,  -- новая связь с ответственным
    status VARCHAR(20) NOT NULL CHECK (status IN ('CREATED', 'PUBLISHED', 'CLOSED')),
    version INTEGER NOT NULL DEFAULT 1,  -- поле для хранения версии
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
