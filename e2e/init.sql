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

CREATE TABLE contacts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    nin VARCHAR(128) NOT NULL,
    dob DATE NULL,
    phone VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    profile_id BIGINT UNSIGNED NOT NULL,
    `type` VARCHAR(255) NOT NULL DEFAULT '',
    INDEX idx_contacts_profile_id (profile_id),
    CONSTRAINT fk_contacts_profile_id FOREIGN KEY (profile_id) REFERENCES profile(id) ON DELETE CASCADE
);
