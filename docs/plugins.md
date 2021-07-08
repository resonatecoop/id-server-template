## Plugins

This server is easily extended or modified through the use of plugins. Four services, [health](https://github.com/RichardKnop/go-oauth2-server/tree/master/health), [oauth](https://github.com/RichardKnop/go-oauth2-server/tree/master/oauth), [session](https://github.com/RichardKnop/go-oauth2-server/tree/master/session) and [web](https://github.com/RichardKnop/go-oauth2-server/tree/master/web) are available for modification.

In order to implement a plugin:
1. Create your own interface that implements all of methods of the service you are replacing.
2. Modify `cmd/run_server.go` to use your service by calling the `session.Use[service-you-are-replaceing]Service(yourCustomService.NewService())` before the services are initialized via `services.Init(cnf, db)`.

For example, to implement an available [redis session storage plugin](https://github.com/adam-hanna/redis-sessions):

~~~go
// $ go get https://github.com/adam-hanna/redis-sessions
//
// cmd/run_server.go
import (
    ...
    "github.com/adam-hanna/redis-sessions/redis"
    ...
)

// RunServer runs the app
func RunServer(configBackend string) error {
    ...

    // configure redis for session store
    sessionSecrets := make([][]byte, 1)
    sessionSecrets[0] = []byte(cnf.Session.Secret)
    redisConfig := redis.ConfigType{
        Size:           10,
        Network:        "tcp",
        Address:        ":6379",
        Password:       "",
        SessionSecrets: sessionSecrets,
    }

    // start the services
    services.UseSessionService(redis.NewService(cnf, redisConfig))
    if err := services.InitServices(cnf, db); err != nil {
        return err
    }
    defer services.CloseServices()

    ...
}
~~~