# Start from Ubuntu base image
FROM ubuntu:22.04

# Install prerequisites
RUN apt-get update && apt-get install -y \
    wget gnupg2 curl apt-transport-https openjdk-17-jdk \
    && rm -rf /var/lib/apt/lists/*

# Add Neo4j GPG key and repository
RUN wget -O - https://debian.neo4j.com/neotechnology.gpg.key | apt-key add - \
    && echo "deb https://debian.neo4j.com stable 5" > /etc/apt/sources.list.d/neo4j.list

# Install Neo4j
RUN apt-get update && apt-get install -y neo4j=1:5.12.0 \
    && rm -rf /var/lib/apt/lists/*

# Environment defaults (can be overridden at runtime)
ENV NEO4J_AUTH=neo4j/neo4j123 \
    NEO4J_server_memory_pagecache_size=1G \
    NEO4J_server_memory_heap_initial__size=1G \
    NEO4J_server_memory_heap_max__size=1G \
    NEO4J_server_security_procedures_unrestricted=gds.*,apoc.* \
    NEO4J_server_directories_data=/neo4j_data \
    NEO4J_server_directories_logs=/neo4j_logs \
    NEO4J_server_directories_import=/var/lib/neo4j/import \
    NEO4J_server_directories_plugins=/neo4j_plugins \
    NEO4J_server_default__listen__address=0.0.0.0 \
    NEO4J_server_bolt_listen__address=0.0.0.0:7687 \
    NEO4J_server_http_listen__address=0.0.0.0:7474

# Create data directories
RUN mkdir -p /neo4j_data /neo4j_logs /neo4j_plugins \
    && chown -R root:root /neo4j_data /neo4j_logs /neo4j_plugins \
    && chmod -R 755 /neo4j_data /neo4j_logs /neo4j_plugins

# Configure Neo4j to listen on all interfaces
RUN echo "server.default_listen_address=0.0.0.0" >> /etc/neo4j/neo4j.conf \
    && echo "server.bolt.listen_address=0.0.0.0:7687" >> /etc/neo4j/neo4j.conf \
    && echo "server.http.listen_address=0.0.0.0:7474" >> /etc/neo4j/neo4j.conf \
    && echo "server.https.listen_address=0.0.0.0:7473" >> /etc/neo4j/neo4j.conf

# Entrypoint script to handle NEO4J_AUTH like the official image
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
if [ -n "${NEO4J_AUTH:-}" ]; then\n\
  if [ "${NEO4J_AUTH}" = "none" ]; then\n\
    echo "Disabling authentication..."\n\
    echo "dbms.security.auth_enabled=false" >> /etc/neo4j/neo4j.conf\n\
  elif [[ "${NEO4J_AUTH}" == neo4j/* ]]; then\n\
    password="${NEO4J_AUTH#neo4j/}"\n\
    if [ ! -d "${NEO4J_server_directories_data}/databases" ]; then\n\
      echo "Setting initial password..."\n\
      neo4j-admin dbms set-initial-password "$password" || true\n\
    fi\n\
  fi\n\
fi\n\
\n\
exec neo4j console' > /docker-entrypoint.sh \
    && chmod +x /docker-entrypoint.sh

# Expose ports
EXPOSE 7474 7687

# Volumes for persistence
VOLUME ["/neo4j_data", "/neo4j_logs", "/var/lib/neo4j/import", "/neo4j_plugins"]

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:7474 || exit 1

# Use entrypoint
CMD ["/docker-entrypoint.sh"]
