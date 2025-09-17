#!/bin/bash
set -e

# Logging function
log() {
    echo "[$(date +%Y-%m-%d\ %H:%M:%S)] $1: $2"
}

# Function to restore PostgreSQL from backup file (based on init.sh)
restore_postgres() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ] || [ ! -f "$backup_file" ]; then
        log "ERROR" "Backup file not found: $backup_file"
        return 1
    fi
    
    log "INFO" "Using backup file: $(basename "$backup_file")"
    
    # Extract backup file
    local temp_dir=$(mktemp -d)
    local backup_name="nexoan.sql"
    
    log "INFO" "Extracting backup file..."
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Copy backup to PostgreSQL backup directory
    log "INFO" "Copying backup to container..."
    cp "$temp_dir/$backup_name" "/var/lib/postgresql/backup/"
    # File is already owned by choreo user since we're running as choreo user
    
    # Check what was actually created
    log "INFO" "Checking backup structure in container..."
    ls -la "/var/lib/postgresql/backup/"
    
    # Restore database
    log "INFO" "Restoring PostgreSQL database..."
    if /usr/lib/postgresql/16/bin/psql -U postgres -d nexoan -f "/var/lib/postgresql/backup/$backup_name"; then
        log "SUCCESS" "PostgreSQL restore completed successfully"
        
        # Verify what was restored
        log "INFO" "Verifying restored database..."
        /usr/lib/postgresql/16/bin/psql -U postgres -d nexoan -c "\dt"
        
        # Clean up
        rm -f "/var/lib/postgresql/backup/$backup_name"
        rm -rf "$temp_dir"
        
    else
        log "ERROR" "PostgreSQL restore failed"
        rm -f "/var/lib/postgresql/backup/$backup_name"
        rm -rf "$temp_dir"
        return 1
    fi
}

# Function to download and extract files from GitHub archive (based on init.sh)
download_github_archive() {
    local version="$1"
    local extract_dir="$2"
    local github_repo="${GITHUB_BACKUP_REPO:-LDFLK/data-backups}"
    
    log "INFO" "Downloading GitHub archive for version $version..."
    
    # Download the archive
    local archive_url="https://github.com/$github_repo/archive/refs/tags/$version.zip"
    local archive_file="$extract_dir/archive.zip"
    
    if wget -q "$archive_url" -O "$archive_file"; then
        log "SUCCESS" "Downloaded archive for version $version"
        
        # Extract the archive
        if unzip -q "$archive_file" -d "$extract_dir"; then
            log "SUCCESS" "Extracted archive"
            rm -f "$archive_file"  # Clean up archive file
            return 0
        else
            log "ERROR" "Failed to extract archive"
            return 1
        fi
    else
        log "ERROR" "Failed to download archive for version $version"
        return 1
    fi
}

# Function to restore from GitHub backup (based on init.sh)
restore_from_github() {
    local version="${BACKUP_VERSION:-0.0.1}"
    local environment="${BACKUP_ENVIRONMENT:-development}"
    
    log "INFO" "Restoring from GitHub version: $version"
    
    # Create temporary directory for downloads
    local temp_dir=$(mktemp -d)
    
    # Download the entire archive
    if ! download_github_archive "$version" "$temp_dir"; then
        log "ERROR" "Failed to download GitHub archive for version $version"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Set the extracted directory path
    local archive_dir="$temp_dir/data-backups-$version"
    
    # Download and restore PostgreSQL
    log "INFO" "Processing PostgreSQL backup..."
    local postgres_file="$archive_dir/nexoan/version/$version/$environment/postgres/nexoan.tar.gz"
    if [ -f "$postgres_file" ]; then
        if restore_postgres "$postgres_file"; then
            log "SUCCESS" "PostgreSQL restored successfully"
            rm -rf "$temp_dir"
            return 0
        else
            log "ERROR" "PostgreSQL restore failed"
            rm -rf "$temp_dir"
            return 1
        fi
    else
        log "ERROR" "PostgreSQL backup not found: $postgres_file"
        rm -rf "$temp_dir"
        return 1
    fi
}

# Clean up any existing lock files from previous runs
log "INFO" "Cleaning up any existing lock files..."
rm -f /var/lib/postgresql/data/postmaster.pid

# Ensure choreo user has proper permissions (volumes may reset ownership)
log "INFO" "Setting up permissions for choreo user..."
# Set permissions (ownership should already be correct from Dockerfile)
chmod -R 755 /var/lib/postgresql/backup /var/log/postgresql
chmod -R 700 /var/lib/postgresql/data

# Initialize PostgreSQL data directory if it's empty
if [ ! -d "/var/lib/postgresql/data" ] || [ -z "$(ls -A /var/lib/postgresql/data 2>/dev/null)" ]; then
    log "INFO" "Initializing PostgreSQL data directory..."
    # Initialize as choreo user
    /usr/lib/postgresql/16/bin/initdb -D /var/lib/postgresql/data
fi

# Start PostgreSQL in background first
log "INFO" "Starting PostgreSQL in background..."
# Run PostgreSQL as choreo user
/usr/lib/postgresql/16/bin/postgres -D /var/lib/postgresql/data &
POSTGRES_PID=$!

# Wait for PostgreSQL to be ready
log "INFO" "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if /usr/lib/postgresql/16/bin/pg_isready -U postgres -q; then
        log "INFO" "PostgreSQL is ready!"
        break
    fi
    log "INFO" "Waiting for PostgreSQL... attempt $i/30"
    sleep 2
done

# Create nexoan database if it doesn't exist
log "INFO" "Creating nexoan database if it doesn't exist..."
/usr/lib/postgresql/16/bin/createdb -U postgres nexoan 2>/dev/null || true

# Check if restore is needed
if [ "${RESTORE_FROM_GITHUB:-false}" = "true" ]; then
    # Check if nexoan database has any tables
    table_count=$(/usr/lib/postgresql/16/bin/psql -U postgres -d nexoan -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' \n' || echo "0")
    
    # Ensure table_count is a valid number
    if ! [[ "$table_count" =~ ^[0-9]+$ ]]; then
        table_count=0
    fi
    
    if [ "$table_count" -eq 0 ]; then
        log "INFO" "nexoan database is empty, starting GitHub restore..."
        restore_from_github || log "WARNING" "GitHub restore failed, continuing with empty database"
    else
        log "INFO" "nexoan database already has $table_count tables, skipping restore"
    fi
fi

# Stop the background PostgreSQL gracefully
log "INFO" "Stopping background PostgreSQL..."
/usr/lib/postgresql/16/bin/pg_ctl stop -D /var/lib/postgresql/data -m smart || kill $POSTGRES_PID 2>/dev/null || true
sleep 3

# Clean up any remaining lock files
rm -f /var/lib/postgresql/data/postmaster.pid

# Start PostgreSQL in foreground as choreo user
log "INFO" "Starting PostgreSQL in foreground mode..."
exec /usr/lib/postgresql/16/bin/postgres -D /var/lib/postgresql/data
