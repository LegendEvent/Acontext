# Database Migrations

This directory contains SQL migration scripts for the Acontext database schema.

## How to Apply Migrations

### Development Environment

Connect to your PostgreSQL database and run the migration:

```bash
psql -h 127.0.0.1 -p 15437 -U acontext -d acontext -f 001_block_reference_set_null.sql
# (docker compose host-export port; container-internal port is 5432)
```

Or using the docker compose setup:

```bash
docker exec -i acontext-server-pg psql -U acontext -d acontext < 001_block_reference_set_null.sql
```

### Production Environment

Always backup your database before applying migrations:

```bash
# Backup
pg_dump -h <host> -U <user> -d <database> > backup_$(date +%Y%m%d_%H%M%S).sql

# Apply migration
psql -h <host> -U <user> -d <database> -f migrations/001_block_reference_set_null.sql
```

## Migration List

| ID  | File                               | Description                                             | Date       |
| --- | ---------------------------------- | ------------------------------------------------------- | ---------- |
| 001 | `001_block_reference_set_null.sql` | Change BlockReference foreign key to SET NULL on delete | 2025-11-04 |

## Migration 001: Block Reference SET NULL

**What it does:**
- Changes the `block_references.reference_block_id` foreign key constraint from CASCADE to SET NULL
- Makes the `reference_block_id` column nullable
- When a referenced block is deleted, the BlockReference record persists with `reference_block_id = NULL` instead of being deleted

**Why:**
- Preserves reference blocks when their target is deleted
- Allows handling of "broken references" in the application
- Prevents unexpected deletion of reference blocks

**Impact:**
- No data loss
- Existing BlockReference records remain unchanged
- Only affects future delete operations on referenced blocks

