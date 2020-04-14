# Qnode
![badge](https://action-badges.now.sh/lunfardo314/goshimmer?action=test)  
This repo uses goshimmer in a very stripped down mode. Most of plugins are disabled.

## Test
The testing configuration consists of 4 goshimmer nodes, running on the same machine on different directories with different config.json files.  

## Setup
Execute the following command on `goshimmer` root directory.
```
$ docker-compose -f docker-compose.qnode.yml up
```

## Stop
Execute the following command on `goshimmer` root directory.
```
$ docker-compose -f docker-compose.qnode.yml down
```

## Abstract
When 4 nodes are running, it is possible to run following programs from tools directory.
- newassembly - It creates smart contract data record in each of goshimmer instances
- newdks - it creates distributed key shares and BLS addresses for testing. They are distributed among 4 instances/nodes
- newconfig - it creates specific configuration data of the committee with those secret shares

When assembly, keys and configurations are in the database of the goshimmer instances, we can run PoC.  
In the current directory `goshimmer\plugins\qnode\tools\mocknode` run following command.  
```
$ ./mocknode mocknode.json
```
it starts value tangle emulator and web server for the FairRulette smart contract.  
To access fair roulette PoC type in browser [http://localhost:2000](http://localhost:2000)

