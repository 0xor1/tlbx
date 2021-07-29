trees
=====

[Live app](https://task-trees.com)

A project management app where tasks are stored in tree structures, the time and cost values estimated and incurred on each task
are rolled up the task tree to the root node, so each node contains a full summary of all the time and cost estimates and incurred
values at all times, making issues with large scale project management more visible.

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
>>>>>>> migrate-to-tlbx-repo
