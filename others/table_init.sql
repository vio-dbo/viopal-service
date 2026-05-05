
-- ============================================
-- TABLE: roles
-- ============================================
CREATE TABLE roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- ============================================
-- TABLE: users
-- ============================================
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role_id BIGINT NOT NULL,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- ============================================
-- TABLE: payment_method
-- ============================================
CREATE TABLE payment_method (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- ============================================
-- TABLE: merchant
-- ============================================
CREATE TABLE merchant (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    bussiness_name VARCHAR(150) NOT NULL,
    phone_number VARCHAR(30),
    balance DECIMAL(18,2) NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ============================================
-- TABLE: top_up
-- ============================================
CREATE TABLE top_up (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ref_number VARCHAR(100) NOT NULL UNIQUE,
    merchant_id BIGINT NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    status VARCHAR(30) NOT NULL,
    request_at DATETIME NOT NULL,
    processed_by BIGINT NULL,
    processed_at DATETIME NULL,
    FOREIGN KEY (merchant_id) REFERENCES merchant(id),
    FOREIGN KEY (processed_by) REFERENCES users(id)
);

-- ============================================
-- TABLE: top_up_logs
-- ============================================
CREATE TABLE top_up_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    top_up_id BIGINT NOT NULL,
    event VARCHAR(100) NOT NULL,
    actor_type VARCHAR(50),
    actor_id BIGINT,
    old_status VARCHAR(30),
    new_status VARCHAR(30),
    balance_before DECIMAL(18,2),
    balance_after DECIMAL(18,2),
    created_at DATETIME NOT NULL,
    FOREIGN KEY (top_up_id) REFERENCES top_up(id)
);

-- ============================================
-- TABLE: invoces
-- ============================================
CREATE TABLE invoices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    merchant_id BIGINT NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    status VARCHAR(30) NOT NULL,
    due_date DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    cust_name VARCHAR(100),
    cust_email VARCHAR(150),
    description TEXT,
    payment_link_token VARCHAR(255),
    FOREIGN KEY (merchant_id) REFERENCES merchant(id)
);

-- ============================================
-- TABLE: payment_intent
-- ============================================
CREATE TABLE payment_intent (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    payment_intent_number VARCHAR(100) NOT NULL UNIQUE,
    invoice_id BIGINT NOT NULL,
    payment_method_id BIGINT NOT NULL,
    payed_by VARCHAR(100),
    approved_by BIGINT NULL,
    approved_at DATETIME NULL,
    failure_reason VARCHAR(255),
    expired_at DATETIME NULL,
    status VARCHAR(30) NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (invoice_id) REFERENCES invoces(id),
    FOREIGN KEY (payment_method_id) REFERENCES payment_method(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
);

-- ============================================
-- TABLE: payment_logs
-- ============================================
CREATE TABLE payment_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    payment_intent_id BIGINT NOT NULL,
    event VARCHAR(100),
    actor_type VARCHAR(50),
    actor_id BIGINT,
    old_status VARCHAR(30),
    new_status VARCHAR(30),
    meta_data JSON,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (payment_intent_id) REFERENCES payment_intent(id)
);

-- ============================================
-- TABLE: refunds
-- ============================================
CREATE TABLE refunds (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    refund_number VARCHAR(100) NOT NULL UNIQUE,
    reason VARCHAR(255),
    payment_intent_id BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    merchant_id BIGINT NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    status VARCHAR(30) NOT NULL,
    request_by BIGINT,
    approved_by BIGINT,
    approved_at DATETIME,
    processed_by BIGINT,
    processed_at DATETIME,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (payment_intent_id) REFERENCES payment_intent(id),
    FOREIGN KEY (invoice_id) REFERENCES invoces(id),
    FOREIGN KEY (merchant_id) REFERENCES merchant(id)
);

-- ============================================
-- TABLE: refund_logs
-- ============================================
CREATE TABLE refund_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    refund_id BIGINT NOT NULL,
    event VARCHAR(100),
    actor_type VARCHAR(50),
    actor_id BIGINT,
    old_status VARCHAR(30),
    new_status VARCHAR(30),
    created_at DATETIME NOT NULL,
    FOREIGN KEY (refund_id) REFERENCES refunds(id)
);

-- ============================================
-- TABLE: merchant_balance_log
-- ============================================
CREATE TABLE merchant_balance_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    merchant_id BIGINT NOT NULL,
    ref_number VARCHAR(100),
    ref_id BIGINT,
    amount DECIMAL(18,2) NOT NULL,
    type VARCHAR(30) NOT NULL,
    balance_before DECIMAL(18,2),
    balance_after DECIMAL(18,2),
    created_at DATETIME NOT NULL,
    FOREIGN KEY (merchant_id) REFERENCES merchant(id)
);