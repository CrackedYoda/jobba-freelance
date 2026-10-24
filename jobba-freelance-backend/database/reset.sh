#!/usr/bin/env bash
# Drops and recreates the local schema, then applies migrations and seed.sql.
set -euo pipefail
cd "$(dirname "$0")"
PSQL=(docker exec -i jobba-freelance-db psql -v ON_ERROR_STOP=1 -U jobba -d jobba_freelance)
"${PSQL[@]}" -c 'drop schema public cascade; create schema public;'
for f in migrations/*.sql; do "${PSQL[@]}" < "$f"; done
"${PSQL[@]}" < seed.sql
if [[ "${1:-}" == "--demo" ]]; then "${PSQL[@]}" < demo.sql; fi
