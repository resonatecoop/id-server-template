
Build a Docker image and run the app in a container:

```
docker build -t go-oauth2-server:latest .
docker run -e ETCD_ENDPOINTS=localhost:2379 -p 8080:8080 --name go-oauth2-server go-oauth2-server:latest
```

### Docker Compose

You can use [docker-compose](https://docs.docker.com/compose/) to start the app, postgres, etcd in separate linked containers:

```
docker-compose up
```

During `docker-compose up` process all configuration and fixtures will be loaded. After successful up you can check, that app is running using for example the health check request:

```
curl --compressed -v localhost:8080/v1/health
```
