-- CES Orchestrator Service Database Initialization
-- Create database and user for orchestrator service

-- Create database if it doesn't exist
SELECT 'CREATE DATABASE ces_orchestrator'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ces_orchestrator')\gexec

-- Connect to the new database
\c ces_orchestrator;

-- Create tables will be handled by GORM auto-migration
-- This script just ensures the database exists

-- Grant permissions (optional, GORM will create tables as needed)
-- GRANT ALL PRIVILEGES ON DATABASE ces_orchestrator TO postgres;