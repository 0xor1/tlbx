trees
====

a simple project management app which stores tasks in trees.

## run

install tmux, go, node/npm and docker/docker-compose then `./bin/run trees`. to kill the development
 services tmux cmd `Ctrl+b &` then `y` to confirm will kill everything.

## Note

This is a port of [a more advanced complex version of trees](https://github.com/0xor1/deprecated-trees) which supports:

* Group accounts
    * For more formal settings group accounts let users be added with more specific
    permissions
* Mulitple data centers:
    * Customers can decide where to host their projects data
    * New geo locations can be added very easily at any time
* DB sharding
    * If the application becomes popular DBs can be sharded to 
    provide scaling

Why port a simpler version? for that very reason, to make it simpler
to setup and configure and develop on with less technical considerations so it's easier to demo and explain.