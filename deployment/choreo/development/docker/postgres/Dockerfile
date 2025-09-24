# PostgreSQL Server Dockerfile
FROM postgres:16

# Set environment variables
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=nexoan

# Create additional directories for backups
RUN mkdir -p /var/lib/postgresql/backup && \
    chown -R postgres:postgres /var/lib/postgresql/backup && \
    chmod -R 755 /var/lib/postgresql/backup

# Define volumes for data persistence
VOLUME ["/var/lib/postgresql/data", "/var/lib/postgresql/backup"]

# Expose ports
EXPOSE 5432

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=5 \
    CMD pg_isready -U postgres || exit 1

# Default command (inherited from base image)
