# Qnode Plugin
## Setup
Copy the dependencies file to current directory and execute `docker-compose`.
```
$ cp ../../go.mod .
$ docker-compose up
```
The `qnode` containers will be listening on ports 8080, 8081, 8082, 8083  
If you want to stop container, execute the following command.
```
$ docker-compose down
```
## Loadmap
There are four steps to complete qnode project.
- Finish developing of the core
- Separate qnode from goshimmer
- Complete qnode api server
- Merge qnode api into goshimmer
