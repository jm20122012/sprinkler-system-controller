#!/bin/bash

source ../.env

echo "Generating migration diff file..."
atlas migrate diff --dir "file://migrations" --to "file://schema.sql" --dev-url "docker://postgres/16"

echo "Applying migration to database..."
atlas migrate apply --url "postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@localhost:5432/sprinkler_controller?search_path=public&sslmode=disable"

