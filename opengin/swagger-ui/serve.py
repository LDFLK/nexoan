import http.server
import socketserver
import os
import logging

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

PORT = 8082

# Get the absolute path to the design directory
DESIGN_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
logger.info(f"Design directory: {DESIGN_DIR}")

class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DESIGN_DIR, **kwargs)
    
    def do_GET(self):
        logger.info(f"Received request for: {self.path}")
        
        # If the path is / or /swagger-ui/, serve index.html
        if self.path in ['/', '/swagger-ui/', '/swagger-ui/index.html']:
            self.path = '/swagger-ui/index.html'
            logger.info(f"Redirecting to: {self.path}")
        
        # Check if the file exists
        file_path = os.path.join(DESIGN_DIR, self.path.lstrip('/'))
        if os.path.exists(file_path):
            logger.info(f"File exists: {file_path}")
        else:
            logger.warning(f"File not found: {file_path}")
        
        return super().do_GET()

with socketserver.TCPServer(("", PORT), Handler) as httpd:
    logger.info(f"Serving Swagger UI at http://localhost:{PORT}")
    logger.info(f"Root directory: {DESIGN_DIR}")
    httpd.serve_forever() 