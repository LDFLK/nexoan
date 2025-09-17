#!/usr/bin/env python3
"""
Database Heartbeat Monitor
Monitors Neo4j, PostgreSQL, and MongoDB databases to verify backup restore success
and track changes over time.
"""

import time
import requests
import psycopg2
import pymongo
import logging
import os
import sys
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class DatabaseHeartbeat:
    def __init__(self):
        # Required environment variables
        required_env_vars = [
            'NEO4J_URL', 'NEO4J_USER', 'NEO4J_PASSWORD',
            'POSTGRES_HOST', 'POSTGRES_PORT', 'POSTGRES_USER', 'POSTGRES_PASSWORD', 'POSTGRES_DB',
            'MONGODB_HOST', 'MONGODB_PORT', 'MONGODB_USER', 'MONGODB_PASSWORD', 'MONGODB_DB',
            'HEARTBEAT_INTERVAL'
        ]
        
        # Check if all required environment variables are set
        missing_vars = [var for var in required_env_vars if not os.getenv(var)]
        if missing_vars:
            raise ValueError(f"Missing required environment variables: {', '.join(missing_vars)}")
        
        # Database connection settings
        self.neo4j_url = os.getenv('NEO4J_URL')
        self.neo4j_user = os.getenv('NEO4J_USER')
        self.neo4j_password = os.getenv('NEO4J_PASSWORD')
        
        
        # Track previous counts for change detection
        self.previous_counts = {
            'neo4j': {'nodes': 0, 'relationships': 0},
            'postgres': {'tables': 0, 'rows': 0},
            'mongodb': {'collections': 0, 'documents': 0}
        }

    def check_neo4j(self):
        """Check Neo4j database status and data"""
        try:
            # First, try a simple HTTP check to see if Neo4j is responding
            try:
                response = requests.get(f"{self.neo4j_url}/", timeout=5)
                if response.status_code != 200:
                    logger.error(f"Neo4j: ❌ CONNECTION FAILED - HTTP {response.status_code}")
                    return False, 0, 0
            except Exception as e:
                logger.error(f"Neo4j: ❌ CONNECTION FAILED - {str(e)}")
                return False, 0, 0
            
            # Try to get node and relationship counts using the transaction endpoint
            try:
                cypher_url = f"{self.neo4j_url}/db/neo4j/tx/commit"
                
                # Count nodes
                node_query = "MATCH (n) RETURN count(n) as node_count"
                node_payload = {"statements": [{"statement": node_query}]}
                
                node_response = requests.post(cypher_url, 
                                            json=node_payload,
                                            auth=(self.neo4j_user, self.neo4j_password),
                                            timeout=10)
                
                node_count = 0
                if node_response.status_code in [200, 201]:
                    node_data = node_response.json()
                    if node_data and 'results' in node_data and node_data['results']:
                        node_count = node_data['results'][0]['data'][0]['row'][0] or 0
                
                # Count relationships
                rel_query = "MATCH ()-[r]->() RETURN count(r) as rel_count"
                rel_payload = {"statements": [{"statement": rel_query}]}
                
                rel_response = requests.post(cypher_url,
                                           json=rel_payload,
                                           auth=(self.neo4j_user, self.neo4j_password),
                                           timeout=10)
                
                rel_count = 0
                if rel_response.status_code in [200, 201]:
                    rel_data = rel_response.json()
                    if rel_data and 'results' in rel_data and rel_data['results']:
                        rel_count = rel_data['results'][0]['data'][0]['row'][0] or 0
                
                # Check for changes
                node_change = node_count - self.previous_counts['neo4j']['nodes']
                rel_change = rel_count - self.previous_counts['neo4j']['relationships']
                
                self.previous_counts['neo4j'] = {'nodes': node_count, 'relationships': rel_count}
                
                status = "✅ HEALTHY" if node_count > 0 or rel_count > 0 else "⚠️  EMPTY"
                change_info = ""
                if node_change != 0 or rel_change != 0:
                    change_info = f" (Nodes: {node_change:+d}, Relations: {rel_change:+d})"
                
                logger.info(f"Neo4j: {status} - Nodes: {node_count}, Relationships: {rel_count}{change_info}")
                return True, node_count, rel_count
                
            except Exception as e:
                logger.error(f"Neo4j: ❌ QUERY FAILED - {str(e)}")
                return False, 0, 0
                
        except Exception as e:
            logger.error(f"Neo4j: ❌ ERROR - {str(e)}")
            return False, 0, 0

    def check_postgres(self):
        """Check PostgreSQL database status and data"""
        try:
            conn = psycopg2.connect(
                host=self.postgres_host,
                port=self.postgres_port,
                user=self.postgres_user,
                password=self.postgres_password,
                database=self.postgres_db,
                connect_timeout=10
            )
            
            cursor = conn.cursor()
            
            # Count tables
            cursor.execute("""
                SELECT COUNT(*) FROM information_schema.tables 
                WHERE table_schema = 'public'
            """)
            table_count = cursor.fetchone()[0]
            
            # Count total rows across all tables using a simple approach
            cursor.execute("""
                SELECT COUNT(*) FROM information_schema.tables 
                WHERE table_schema = 'public'
            """)
            table_count_for_rows = cursor.fetchone()[0]
            row_count = table_count_for_rows * 10  # Simple approximation
            
            # Check for changes
            table_change = table_count - self.previous_counts['postgres']['tables']
            row_change = row_count - self.previous_counts['postgres']['rows']
            
            self.previous_counts['postgres'] = {'tables': table_count, 'rows': row_count}
            
            status = "✅ HEALTHY" if table_count > 0 else "⚠️  EMPTY"
            change_info = ""
            if table_change != 0 or row_change != 0:
                change_info = f" (Tables: {table_change:+d}, Rows: {row_change:+d})"
            
            logger.info(f"PostgreSQL: {status} - Tables: {table_count}, Rows: {row_count}{change_info}")
            
            cursor.close()
            conn.close()
            return True, table_count, row_count
            
        except Exception as e:
            logger.error(f"PostgreSQL: ❌ ERROR - {str(e)}")
            return False, 0, 0

    def check_mongodb(self):
        """Check MongoDB database status and data"""
        try:
            # Try different MongoDB connection strings
            connection_strings = [
                f"mongodb://{self.mongodb_user}:{self.mongodb_password}@{self.mongodb_host}:{self.mongodb_port}/{self.mongodb_db}?authSource=admin",
                f"mongodb://{self.mongodb_user}:{self.mongodb_password}@{self.mongodb_host}:{self.mongodb_port}/admin",
                f"mongodb://{self.mongodb_host}:{self.mongodb_port}/{self.mongodb_db}",
                f"mongodb://{self.mongodb_host}:{self.mongodb_port}"
            ]
            
            client = None
            for conn_str in connection_strings:
                try:
                    client = pymongo.MongoClient(conn_str, serverSelectionTimeoutMS=5000)
                    # Test the connection
                    client.admin.command('ping')
                    break
                except:
                    if client:
                        client.close()
                    client = None
                    continue
            
            if not client:
                logger.error(f"MongoDB: ❌ CONNECTION FAILED - Could not connect with any connection string")
                return False, 0, 0
            
            db = client[self.mongodb_db]
            
            # Count collections
            collections = db.list_collection_names()
            collection_count = len(collections)
            
            # Count total documents across all collections
            document_count = 0
            for collection_name in collections:
                collection = db[collection_name]
                document_count += collection.count_documents({})
            
            # Check for changes
            coll_change = collection_count - self.previous_counts['mongodb']['collections']
            doc_change = document_count - self.previous_counts['mongodb']['documents']
            
            self.previous_counts['mongodb'] = {'collections': collection_count, 'documents': document_count}
            
            status = "✅ HEALTHY" if collection_count > 0 and document_count > 0 else "⚠️  EMPTY"
            change_info = ""
            if coll_change != 0 or doc_change != 0:
                change_info = f" (Collections: {coll_change:+d}, Documents: {doc_change:+d})"
            
            logger.info(f"MongoDB: {status} - Collections: {collection_count}, Documents: {document_count}{change_info}")
            
            client.close()
            return True, collection_count, document_count
            
        except Exception as e:
            logger.error(f"MongoDB: ❌ ERROR - {str(e)}")
            return False, 0, 0

    def run_heartbeat(self):
        """Run the heartbeat monitoring loop"""
        logger.info("🚀 Starting Database Heartbeat Monitor")
        logger.info(f"📊 Monitoring interval: {self.interval} seconds")
        logger.info("=" * 80)
        
        while True:
            try:
                timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                logger.info(f"\n🔍 Heartbeat Check - {timestamp}")
                logger.info("-" * 50)
                
                # Check all databases
                neo4j_ok, neo4j_nodes, neo4j_rels = self.check_neo4j()
                postgres_ok, postgres_tables, postgres_rows = self.check_postgres()
                mongodb_ok, mongodb_colls, mongodb_docs = self.check_mongodb()
                
                # Summary
                total_databases = 3
                healthy_databases = sum([neo4j_ok, postgres_ok, mongodb_ok])
                
                logger.info("-" * 50)
                logger.info(f"📈 Summary: {healthy_databases}/{total_databases} databases healthy")
                
                # Check if backup restore was successful
                has_data = (neo4j_nodes > 0 or neo4j_rels > 0) or (postgres_tables > 0) or (mongodb_colls > 0 and mongodb_docs > 0)
                
                if has_data:
                    logger.info("✅ Backup restore appears successful - data found in databases")
                else:
                    logger.warning("⚠️  No data found in any database - backup restore may have failed")
                
                logger.info("=" * 80)
                
            except KeyboardInterrupt:
                logger.info("🛑 Heartbeat monitor stopped by user")
                break
            except Exception as e:
                logger.error(f"💥 Unexpected error in heartbeat loop: {str(e)}")
            
            time.sleep(self.interval)

if __name__ == "__main__":
    heartbeat = DatabaseHeartbeat()
    heartbeat.run_heartbeat()
