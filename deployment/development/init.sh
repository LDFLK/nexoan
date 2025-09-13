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
        "execute")
            execute
            ;;
        "finalize")
            finalize
            ;;
        "help"|*)
            echo "Usage: $0 {setup|setup_neo4j|run_neo4j|execute|finalize|help}"
            ;;
    esac
}

# Run main function with all arguments
main "$@"