CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    notification_subject VARCHAR(50) NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    status VARCHAR(10) NOT NULL,
    subject VARCHAR(255),
    sent_at TIMESTAMP,
    failed_reason TEXT,
    retry_count INT DEFAULT 0,
    event_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_email ON notifications(email);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
