CREATE DATABASE IF NOT EXISTS profile;
USE profile;

CREATE TABLE profile (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE INDEX idx_profile_user_id (user_id),
    INDEX idx_profile_email (email)
);
