#!/bin/bash
set -e

echo "loading schema for tests..."
echo "drop database if exists bezosphere_test; create database bezosphere_test" | psql -U postgres -h postgres
migrate -url $BEZOSPHERE_POSTGRES_CONN -path ./migrations up
