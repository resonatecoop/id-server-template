## Session Storage

By default, this server implements in-memory, cookie sessions via [gorilla sessions](https://github.com/gorilla/sessions).

However, because the session service can be replaced via a plugin, any of the available [gorilla sessions store implementations](https://github.com/gorilla/sessions#store-implementations) can be wrapped by `session.ServiceInterface`.

## Dependencies

Since Go 1.11, a new recommended dependency management system is via [modules](https://github.com/golang/go/wiki/Modules).

This is one of slight weaknesses of Go as dependency management is not a solved problem. Previously Go was officially recommending to use the [dep tool](https://github.com/golang/dep) but that has been abandoned now in favor of modules.

## Variables

Live environments require the following etcd environment variables

* `ETCD_ENDPOINTS`
* `ETCD_CERT_FILE`
* `ETCD_KEY_FILE`
* `ETCD_CA_FILE`
* `ETCD_CONFIG_PATH`