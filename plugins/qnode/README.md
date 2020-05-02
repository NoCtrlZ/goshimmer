# Qnode Plugin
## Setup
Copy the default config file to `qnode-test` directory and run qnode api server.
```
$ cp ../../config.default.json ./config.json
$ go run qnode-test/main.go
```

## Docker
Execute the following command on `root` directory.
```
$ docker-compose -f docker-compose.qnode.yml up
```
The `qnode` api server will be listening on [localhost:8080](http://localhost/8080/adm/getsclist)  
If you want to stop container, execute the following command.
```
$ docker-compose -f docker-compose.qnode.yml down
```
## Loadmap
There are four steps to complete qnode project.
- Finish developing of the core
- Separate qnode from goshimmer
- Complete qnode api server
- Merge qnode api into goshimmer
