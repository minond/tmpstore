# Temporarily stores files and gives you a URL for retrieval

```bash
PORT=8080 go run main.go

curl -X POST -F file=@main.go localhost:8080/upload
# -> {"name":"cxsLCszzAYCmeSpndBxvrCRUKWLeBqtb"}

curl localhost:8080/get/cxsLCszzAYCmeSpndBxvrCRUKWLeBqtb
```
