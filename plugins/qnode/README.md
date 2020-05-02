# Qnode Plugin
## Setup
Copy the default config file to `qnode-test` directory and run qnode api server.
```
$ cp ../../config.default.json ./config.json
$ go run qnode-test/main.go
```

## Loadmap
There are four steps to complete qnode project.
- Finish developing of the core
- Separate qnode from goshimmer
- Complete qnode api server
- Merge qnode api into goshimmer