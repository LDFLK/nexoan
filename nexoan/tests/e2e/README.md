## Setup 

```bash
cd design/ingestion-api && source .env && bal run
cd ../../
cd design/query-api && source .env && bal run
python basic_core_tests.py
python basic_query_tests.py
```