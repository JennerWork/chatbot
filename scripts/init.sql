-- 创建客户表
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    salt VARCHAR(32) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_deleted_at ON customers(deleted_at);

-- 创建会话表
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    status VARCHAR(20) NOT NULL,
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sessions_customer_id ON sessions(customer_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_last_active_at ON sessions(last_active_at);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);

-- 创建消息表
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    session_id INTEGER NOT NULL REFERENCES sessions(id),
    sender VARCHAR(20) NOT NULL,
    seq INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_messages_customer_id ON messages(customer_id);
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_seq ON messages(seq);
CREATE INDEX idx_messages_deleted_at ON messages(deleted_at);

-- 创建反馈表
CREATE TABLE feedbacks (
    id SERIAL PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    message_id INTEGER REFERENCES messages(id),
    session_id INTEGER NOT NULL REFERENCES sessions(id),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_feedbacks_customer_id ON feedbacks(customer_id);
CREATE INDEX idx_feedbacks_message_id ON feedbacks(message_id);
CREATE INDEX idx_feedbacks_session_id ON feedbacks(session_id);
CREATE INDEX idx_feedbacks_status ON feedbacks(status);
CREATE INDEX idx_feedbacks_deleted_at ON feedbacks(deleted_at);

-- 创建触发器函数来自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为每个表添加自动更新 updated_at 的触发器
CREATE TRIGGER update_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at
    BEFORE UPDATE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_feedbacks_updated_at
    BEFORE UPDATE ON feedbacks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 