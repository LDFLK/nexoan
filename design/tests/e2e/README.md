## Setup 

```bash
cd design/update-api && source .env && bal run
cd ../../
cd design/query-api && source .env && bal run
python basic_crud_tests.py
python basic_query_tests.py
```