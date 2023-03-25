# Compilation

To re-build image:

```
cue fmt internal/swagger/swagger.cue
go generate ./...
docker-compose build
```

Swagger URL: http://localhost:8080/swagger

Look inside [docker-compose.yaml](docker-compose.yaml) for an example of the connection string to use to connect to the local oracle db.
