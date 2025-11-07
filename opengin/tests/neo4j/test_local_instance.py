from neo4j import GraphDatabase
import os

# Neo4j connection details
NEO4J_URI = os.getenv("NEO4J_TESTING_DB_URI", "")
NEO4J_USER = os.getenv("NEO4J_TESTING_USERNAME", "")
NEO4J_PASSWORD = os.getenv("NEO4J_TESTING_PASSWORD", "")

def test_local_instance():
    try:
        driver = GraphDatabase.driver(NEO4J_URI, auth=(NEO4J_USER, NEO4J_PASSWORD))

        with driver.session() as session:
            result = session.run("RETURN 'Connected to Neo4j' AS message")
            message = result.single()["message"]
            assert message == "Connected to Neo4j", f"Expected 'Connected to Neo4j', got '{message}'"
            print(f"Successfully connected to Neo4j: {message}")

        driver.close()
    except Exception as e:
        print(f"Error connecting to Neo4j: {str(e)}")

if __name__ == "__main__":
    test_local_instance()
