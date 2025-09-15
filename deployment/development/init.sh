#!/bin/bash

# Neo4j Backup Initialization Script
# Simple setup, execute, and finalize functions

set -e

# Configuration
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Setup function
setup() {
    log "INFO" "Setting up backup environment..."
    source ../../configs/backup.env
    echo "NEO4J_BACKUP_DIR: $NEO4J_BACKUP_DIR"
    echo "POSTGRES_BACKUP_DIR: $POSTGRES_BACKUP_DIR"
    echo "MONGODB_BACKUP_DIR: $MONGODB_BACKUP_DIR"
    # Add your setup logic here
    log "SUCCESS" "Setup completed"
}

# Setup Neo4j - Build the Docker image
setup_neo4j() {
    log "INFO" "Building Neo4j-backup service using docker-compose..."
    docker-compose -f ../../docker-compose.yml build neo4j
    log "SUCCESS" "Neo4j-backup service built successfully"
}

# Run Neo4j container
run_neo4j() {
    log "INFO" "Loading environment variables..."
    source ../../configs/backup.env
    
    log "INFO" "Starting Neo4j-backup service using docker-compose..."
    docker-compose -f ../../docker-compose.yml up -d neo4j
    log "SUCCESS" "Neo4j-backup service started successfully"
}

# MongoDB Backup Functions
backup_mongodb() {
    log "INFO" "Starting MongoDB backup..."
    source ../../configs/backup.env
    
    local backup_dir="${MONGODB_BACKUP_DIR:-./backups/mongodb}"
    local backup_file="nexoan"
    
    # Create backup directory
    mkdir -p "$backup_dir"
    
    log "INFO" "Creating MongoDB dump..."
    echo "MONGODB_USERNAME: $MONGODB_USERNAME"
    echo "MONGODB_PASSWORD: $MONGODB_PASSWORD"
    echo "MONGODB_DATABASE: $MONGODB_DATABASE"
    echo "backup_dir: $backup_dir"
    echo "backup_file: $backup_file"
    
    # Ensure backup directory exists in container
    log "INFO" "Creating backup directory in container..."
    docker exec mongodb mkdir -p /data/backup
    
    # Test MongoDB connection first
    log "INFO" "Testing MongoDB connection..."
    if docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
        log "SUCCESS" "MongoDB connection successful"
    else
        log "ERROR" "MongoDB connection failed"
        return 1
    fi
    
    # Check what databases exist
    log "INFO" "Checking available databases..."
    docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db.adminCommand('listDatabases')"
    
    # Check if target database exists and has collections
    log "INFO" "Checking target database: ${MONGODB_DATABASE}"
    docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db = db.getSiblingDB('${MONGODB_DATABASE}'); db.getCollectionNames()"
    
    # Run mongodump and capture output
    log "INFO" "Running mongodump command..."
    mongodump_output=$(docker exec mongodb mongodump --host=localhost:27017 \
        --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} \
        --authenticationDatabase=admin \
        --db=${MONGODB_DATABASE} \
        --out="/data/backup/${backup_file}" 2>&1)
    mongodump_exit_code=$?
    
    log "INFO" "Mongodump output: $mongodump_output"
    log "INFO" "Mongodump exit code: $mongodump_exit_code"
    
    if [ $mongodump_exit_code -eq 0 ]; then
        
        log "SUCCESS" "MongoDB dump command completed"
        
        # Check what was actually created
        log "INFO" "Checking backup directory contents..."
        docker exec mongodb ls -la "/data/backup/"
        
        # Verify backup was created
        log "INFO" "Verifying backup files..."
        if docker exec mongodb test -d "/data/backup/${backup_file}"; then
            docker exec mongodb ls -la "/data/backup/${backup_file}"
        else
            log "WARNING" "Backup directory not found, trying alternative approach..."
            # Try creating the directory first and running mongodump again
            docker exec mongodb mkdir -p "/data/backup/${backup_file}"
            log "INFO" "Retrying mongodump with pre-created directory..."
            if docker exec mongodb mongodump --host=localhost:27017 \
                --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} \
                --authenticationDatabase=admin \
                --db=${MONGODB_DATABASE} \
                --out="/data/backup/${backup_file}" 2>&1; then
                log "SUCCESS" "Retry successful"
                docker exec mongodb ls -la "/data/backup/${backup_file}"
            else
                log "ERROR" "Backup failed even with retry"
                return 1
            fi
        fi
        
        # Copy backup from container to host
        log "INFO" "Copying backup to host..."
        docker cp "mongodb:/data/backup/${backup_file}" "$backup_dir/"
        
        # Create compressed archive
        log "INFO" "Creating compressed archive..."
        cd "$backup_dir"
        tar -czf "nexoan.tar.gz" "$backup_file"
        rm -rf "$backup_file"
        
        # Clean up container backup
        docker exec mongodb rm -rf "/data/backup/${backup_file}"
        
        log "SUCCESS" "MongoDB backup completed: nexoan.tar.gz"
    else
        log "ERROR" "MongoDB backup failed"
        return 1
    fi
}

# MongoDB Restore Functions
restore_mongodb() {
    log "INFO" "Starting MongoDB restore..."
    source ../../configs/backup.env
    
    local backup_dir="${MONGODB_BACKUP_DIR:-./backups/mongodb}"
    
    if [ ! -d "$backup_dir" ]; then
        log "ERROR" "Backup directory not found: $backup_dir"
        return 1
    fi
    
    log "INFO" "Using backup directory: $backup_dir"
    
    # Look for nexoan.tar.gz file
    local backup_file="$backup_dir/nexoan.tar.gz"
    
    if [ ! -f "$backup_file" ]; then
        log "ERROR" "Backup file not found: $backup_file"
        return 1
    fi
    
    log "INFO" "Using backup file: $(basename "$backup_file")"
    
    # Extract backup file
    local temp_dir=$(mktemp -d)
    local backup_name="nexoan"
    
    log "INFO" "Extracting backup file..."
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Copy backup to container
    log "INFO" "Copying backup to container..."
    docker cp "$temp_dir/$backup_name" "mongodb:/data/backup/"
    
    # Check what was actually created
    log "INFO" "Checking backup structure in container..."
    docker exec mongodb find "/data/backup/$backup_name" -type d -name "*" 2>/dev/null || true
    docker exec mongodb ls -la "/data/backup/$backup_name" 2>/dev/null || true
    
    # Use the backup directory directly (mongorestore will handle the database structure)
    local db_path="/data/backup/$backup_name"
    
    log "INFO" "Using backup path: $db_path"
    
    # Restore database
    log "INFO" "Restoring MongoDB database..."
    if docker exec mongodb mongorestore --host=localhost:27017 \
        --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} \
        --authenticationDatabase=admin \
        --drop \
        "$db_path"; then
        
        log "SUCCESS" "MongoDB restore completed successfully"
        
        # Verify what was restored
        log "INFO" "Verifying restored databases..."
        docker exec mongodb mongo --host=localhost:27017 --username=${MONGODB_USERNAME} --password=${MONGODB_PASSWORD} --authenticationDatabase=admin --eval "db.adminCommand('listDatabases')"
        
        # Clean up
        docker exec mongodb rm -rf "/data/backup/$backup_name"
        rm -rf "$temp_dir"
        
    else
        log "ERROR" "MongoDB restore failed"
        rm -rf "$temp_dir"
        return 1
    fi
}

# List MongoDB backups
list_mongodb_backups() {
    source ../../configs/backup.env
    local backup_dir="${MONGODB_BACKUP_DIR:-./backups/mongodb}"
    
    log "INFO" "MongoDB backups in: $backup_dir"
    
    if [ -d "$backup_dir" ]; then
        ls -la "$backup_dir"/*.tar.gz 2>/dev/null || log "WARNING" "No backup files found"
    else
        log "WARNING" "Backup directory does not exist: $backup_dir"
    fi
}

backup_neo4j() {
    log "INFO" "Backing up Neo4j-backup service using docker-compose..."
    docker-compose -f ../../docker-compose.yml down neo4j
    log "SUCCESS" "Neo4j-backup service backed up successfully"
    
    ## Sample Command to backup Neo4j
    ## docker run --rm \
    ##     --volume=/var/lib/docker/volumes/neo4j_data/_data:/data \
    ##     --volume=/Users/username/github/fork/data-backups/nexoan/neo4j/backups:/backups \
    ##     neo4j/neo4j-admin:latest \
    ##     neo4j-admin database dump neo4j --to-path=/backups
}

# Execute function
execute() {
    log "INFO" "Executing backup operations..."
    # Add your execute logic here
    log "SUCCESS" "Execute completed"
}

# Finalize function
finalize() {
    log "INFO" "Finalizing backup process..."
    # Add your finalize logic here
    log "SUCCESS" "Finalize completed"
}

# Main function
main() {
    local command="${1:-help}"
    
    case $command in
        "setup")
            setup
            ;;
        "setup_neo4j")
            setup_neo4j
            ;;
        "run_neo4j")
            run_neo4j
            ;;
        "backup_mongodb")
            backup_mongodb
            ;;
        "restore_mongodb")
            restore_mongodb "$2"
            ;;
        "list_mongodb_backups")
            list_mongodb_backups
            ;;
        "execute")
            execute
            ;;
        "finalize")
            finalize
            ;;
        "help"|*)
            echo "Usage: $0 {setup|setup_neo4j|run_neo4j|backup_mongodb|restore_mongodb|list_mongodb_backups|execute|finalize|help}"
            echo ""
            echo "MongoDB Commands:"
            echo "  backup_mongodb        - Create MongoDB backup"
            echo "  restore_mongodb <file> - Restore MongoDB from backup file"
            echo "  list_mongodb_backups  - List available MongoDB backups"
            ;;
    esac
}

# Run main function with all arguments
main "$@"