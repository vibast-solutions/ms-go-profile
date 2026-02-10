# Profile Service Deployment Guide

This document describes what `profile` needs in development and production.

## 1. Runtime Topology

- Service binary: `profile-service` (command: `serve`)
- Protocols: HTTP + gRPC in the same process
- Default ports:
- HTTP: `8080` (configurable with `HTTP_PORT`)
- gRPC: `9090` (configurable with `GRPC_PORT`)
- External dependencies:
- MySQL: required
- Redis: not used

## 2. Environment Variables

Required:

- `MYSQL_DSN`

Optional (with defaults):

- `HTTP_HOST` (default `0.0.0.0`)
- `HTTP_PORT` (default `8080`)
- `GRPC_HOST` (default `0.0.0.0`)
- `GRPC_PORT` (default `9090`)
- `MYSQL_MAX_OPEN_CONNS` (default `10`)
- `MYSQL_MAX_IDLE_CONNS` (default `5`)
- `MYSQL_CONN_MAX_LIFETIME_MINUTES` (default `30`)
- `LOG_LEVEL` (default `info`)

Example DSN:

- `MYSQL_DSN=user:pass@tcp(mysql-host:3306)/profile?parseTime=true`

## 3. MySQL Requirements

Database:

- name: `profile`

Tables and indexes expected by the service:

```sql
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
    INDEX idx_contacts_profile_id_type (profile_id, `type`),
    CONSTRAINT fk_contacts_profile_id FOREIGN KEY (profile_id) REFERENCES profile(id) ON DELETE CASCADE
);

CREATE TABLE addresses (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    street_name VARCHAR(255) NOT NULL,
    streen_no VARCHAR(128) NOT NULL,
    city VARCHAR(255) NOT NULL,
    county VARCHAR(255) NOT NULL,
    country VARCHAR(255) NOT NULL,
    profile_id BIGINT UNSIGNED NOT NULL,
    postal_code VARCHAR(64) NOT NULL DEFAULT '',
    building VARCHAR(128) NOT NULL DEFAULT '',
    apartment VARCHAR(128) NOT NULL DEFAULT '',
    additional_data VARCHAR(512) NOT NULL DEFAULT '',
    `type` VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_addresses_profile_id_type (profile_id, `type`),
    CONSTRAINT fk_addresses_profile_id FOREIGN KEY (profile_id) REFERENCES profile(id) ON DELETE CASCADE
);

CREATE TABLE companies (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    registration_no VARCHAR(255) NOT NULL,
    fiscal_code VARCHAR(255) NOT NULL,
    profile_id BIGINT UNSIGNED NOT NULL,
    `type` VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_companies_profile_id_type (profile_id, `type`),
    CONSTRAINT fk_companies_profile_id FOREIGN KEY (profile_id) REFERENCES profile(id) ON DELETE CASCADE
);
```

## 4. Development Setup

Recommended local stack:

- MySQL 8.x
- service process (`go run main.go serve` or built binary)

Reference e2e compose:

- `/Users/stefan.balea/projects/microservices-ecosystem/profile/e2e/docker-compose.yml`

## 5. Production Notes

- Use least-privilege DB user on `profile` schema.
- Keep MySQL backups and migration rollout process in place.
- Place TLS/ingress in front of HTTP/gRPC listeners.
- Keep `LOG_LEVEL=info` (or `warn`) in production by default.
