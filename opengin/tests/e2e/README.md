## Setup 

```bash
cd design/ingestion-api && source .env && bal run
cd ../../
cd opengin/read-api && source .env && bal run
python basic_core_tests.py
python basic_read_tests.py
```