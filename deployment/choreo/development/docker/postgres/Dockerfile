# PostgreSQL Server Dockerfile with GitHub backup restore
FROM postgres:16

# Install additional tools for backup restore
RUN apt-get update && apt-get install -y \
    wget unzip \
    && rm -rf /var/lib/apt/lists/*

# Create choreo user and group (required for Choreo platform)
RUN groupadd -g 10014 choreo && \
    useradd -u 10014 -g choreo -s /bin/bash -m choreouser && \
    mkdir -p /home/choreouser/.cache && \
    chown -R choreouser:choreo /home/choreouser && \
    chmod -R 755 /home/choreouser

# Create data directories as root and set ownership to choreo user
RUN mkdir -p /var/lib/postgresql/backup /var/lib/postgresql/data /var/log/postgresql && \
    chown -R 10014:10014 /var/lib/postgresql/backup /var/lib/postgresql/data /var/log/postgresql

# Switch to choreo user (required for Choreo platform)
USER 10014

# Set environment variables
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=nexoan

# GitHub backup restore configuration
ENV GITHUB_BACKUP_REPO=${NEXOAN_GITHUB_BACKUP_REPO:-LDFLK/data-backups} \
    BACKUP_VERSION=${NEXOAN_DB_BACKUP_VERSION:-0.0.1} \
    BACKUP_ENVIRONMENT=${NEXOAN_CHOREO_ENVIRONMENT:-development} \
    RESTORE_FROM_GITHUB=true

# Create entrypoint script in choreo user's home directory
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
# Logging function\n\
log() {\n\
    echo "[$(date +%Y-%m-%d\ %H:%M:%S)] $1: $2"\n\
}\n\
\n\
# Function to restore PostgreSQL from backup file\n\
restore_postgres() {\n\
    local backup_file="$1"\n\
    \n\
    if [ -z "$backup_file" ] || [ ! -f "$backup_file" ]; then\n\
        log "ERROR" "Backup file not found: $backup_file"\n\
        return 1\n\
    fi\n\
    \n\
    log "INFO" "Using backup file: $(basename "$backup_file")"\n\
    \n\
    # Extract backup file\n\
    local temp_dir=$(mktemp -d)\n\
    local backup_name="nexoan.sql"\n\
    \n\
    log "INFO" "Extracting backup file..."\n\
    tar -xzf "$backup_file" -C "$temp_dir"\n\
    \n\
    # Copy backup to PostgreSQL backup directory\n\
    log "INFO" "Copying backup to container..."\n\
    cp "$temp_dir/$backup_name" "/var/lib/postgresql/backup/"\n\
    \n\
    # Restore database\n\
    log "INFO" "Restoring PostgreSQL database..."\n\
    if /usr/lib/postgresql/16/bin/psql -U postgres -d nexoan -f "/var/lib/postgresql/backup/$backup_name"; then\n\
        log "SUCCESS" "PostgreSQL restore completed successfully"\n\
        rm -f "/var/lib/postgresql/backup/$backup_name"\n\
        rm -rf "$temp_dir"\n\
        return 0\n\
    else\n\
        log "ERROR" "PostgreSQL restore failed"\n\
        rm -f "/var/lib/postgresql/backup/$backup_name"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
}\n\
\n\
# Function to download and extract files from GitHub archive\n\
download_github_archive() {\n\
    local version="$1"\n\
    local extract_dir="$2"\n\
    local github_repo="${GITHUB_BACKUP_REPO:-LDFLK/data-backups}"\n\
    \n\
    log "INFO" "Downloading GitHub archive for version $version..."\n\
    \n\
    # Download the archive\n\
    local archive_url="https://github.com/$github_repo/archive/refs/tags/$version.zip"\n\
    local archive_file="$extract_dir/archive.zip"\n\
    \n\
    if wget -q "$archive_url" -O "$archive_file"; then\n\
        log "SUCCESS" "Downloaded archive for version $version"\n\
        \n\
        # Extract the archive\n\
        if unzip -q "$archive_file" -d "$extract_dir"; then\n\
            log "SUCCESS" "Extracted archive"\n\
            rm -f "$archive_file"\n\
            return 0\n\
        else\n\
            log "ERROR" "Failed to extract archive"\n\
            return 1\n\
        fi\n\
    else\n\
        log "ERROR" "Failed to download archive for version $version"\n\
        return 1\n\
    fi\n\
}\n\
\n\
# Function to restore from GitHub backup\n\
restore_from_github() {\n\
    local version="${BACKUP_VERSION:-0.0.1}"\n\
    local environment="${BACKUP_ENVIRONMENT:-development}"\n\
    \n\
    log "INFO" "Restoring from GitHub version: $version"\n\
    \n\
    # Create temporary directory for downloads\n\
    local temp_dir=$(mktemp -d)\n\
    \n\
    # Download the entire archive\n\
    if ! download_github_archive "$version" "$temp_dir"; then\n\
        log "ERROR" "Failed to download GitHub archive for version $version"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
    \n\
    # Set the extracted directory path\n\
    local archive_dir="$temp_dir/data-backups-$version"\n\
    \n\
    # Download and restore PostgreSQL\n\
    log "INFO" "Processing PostgreSQL backup..."\n\
    local postgres_file="$archive_dir/nexoan/version/$version/$environment/postgres/nexoan.tar.gz"\n\
    if [ -f "$postgres_file" ]; then\n\
        if restore_postgres "$postgres_file"; then\n\
            log "SUCCESS" "PostgreSQL restored successfully"\n\
            rm -rf "$temp_dir"\n\
            return 0\n\
        else\n\
            log "ERROR" "PostgreSQL restore failed"\n\
            rm -rf "$temp_dir"\n\
            return 1\n\
        fi\n\
    else\n\
        log "ERROR" "PostgreSQL backup not found: $postgres_file"\n\
        rm -rf "$temp_dir"\n\
        return 1\n\
    fi\n\
}\n\
\n\
# Clean up any existing lock files from previous runs\n\
log "INFO" "Cleaning up any existing lock files..."\n\
rm -f /var/lib/postgresql/data/postmaster.pid\n\
\n\
# Start PostgreSQL in background first\n\
log "INFO" "Starting PostgreSQL in background..."\n\
/usr/lib/postgresql/16/bin/postgres -D /var/lib/postgresql/data &\n\
POSTGRES_PID=$!\n\
\n\
# Wait for PostgreSQL to be ready\n\
log "INFO" "Waiting for PostgreSQL to be ready..."\n\
for i in {1..30}; do\n\
    if /usr/lib/postgresql/16/bin/pg_isready -U postgres -q; then\n\
        log "INFO" "PostgreSQL is ready!"\n\
        break\n\
    fi\n\
    log "INFO" "Waiting for PostgreSQL... attempt $i/30"\n\
    sleep 2\n\
done\n\
\n\
# Check if restore is needed\n\
if [ "${RESTORE_FROM_GITHUB:-false}" = "true" ]; then\n\
    # Check if nexoan database has any tables\n\
    table_count=$(/usr/lib/postgresql/16/bin/psql -U postgres -d nexoan -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '\''public'\'';" 2>/dev/null | tr -d '\'' \n'\'' || echo "0")\n\
    \n\
    # Ensure table_count is a valid number\n\
    if ! [[ "$table_count" =~ ^[0-9]+$ ]]; then\n\
        table_count=0\n\
    fi\n\
    \n\
    if [ "$table_count" -eq 0 ]; then\n\
        log "INFO" "nexoan database is empty, starting GitHub restore..."\n\
        restore_from_github || log "WARNING" "GitHub restore failed, continuing with empty database"\n\
    else\n\
        log "INFO" "nexoan database already has $table_count tables, skipping restore"\n\
    fi\n\
fi\n\
\n\
# Stop the background PostgreSQL gracefully\n\
log "INFO" "Stopping background PostgreSQL..."\n\
/usr/lib/postgresql/16/bin/pg_ctl stop -D /var/lib/postgresql/data -m smart || kill $POSTGRES_PID 2>/dev/null || true\n\
sleep 3\n\
\n\
# Clean up any remaining lock files\n\
rm -f /var/lib/postgresql/data/postmaster.pid\n\
\n\
# Start PostgreSQL in foreground as choreo user\n\
log "INFO" "Starting PostgreSQL in foreground mode..."\n\
exec /usr/lib/postgresql/16/bin/postgres -D /var/lib/postgresql/data' > /home/choreouser/custom-entrypoint.sh \
    && chmod +x /home/choreouser/custom-entrypoint.sh

# Initialize PostgreSQL data directory as choreo user
RUN /usr/lib/postgresql/16/bin/initdb -D /var/lib/postgresql/data

# Start PostgreSQL in background for setup
RUN /usr/lib/postgresql/16/bin/postgres -D /var/lib/postgresql/data &
RUN sleep 10

# Create nexoan database
RUN /usr/lib/postgresql/16/bin/createdb -U postgres nexoan 2>/dev/null || true

# Stop PostgreSQL
RUN /usr/lib/postgresql/16/bin/pg_ctl stop -D /var/lib/postgresql/data -m smart || true


# Define volumes for data persistence
VOLUME ["/var/lib/postgresql/data", "/var/lib/postgresql/backup"]

# Expose ports
EXPOSE 5432

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=5 \
    CMD pg_isready -U postgres || exit 1

# Use custom entrypoint
CMD ["/home/choreouser/custom-entrypoint.sh"]