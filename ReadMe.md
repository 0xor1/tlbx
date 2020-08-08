tlbx
===

## fmt & test

```
./bin/dkr
```
to up the service containers that are used by the apps.
```
./bin/dkr reset
```
to fully rebuild and run all the docker images/containers.
```
./bin/pre
```
to fmt and test everything.

## structure

* `/pkg` - reusable packages
* `/bin` - util scripts to run/build/test
* `/cmd` - executable go programs
* `/sql` - sql schemas
* `/docker` - docker-compose files for dev environment setup

## web apps

tlbx is predominantly for making web apps that follow a similar pattern, to run these apps tmux, go,
 node/npm and docker/docker-compose must be installed then simply `./bin/run <app_name>` e.g.
 `./bin/run todo`. to kill the development services tmux cmd `Ctrl+b &` then `y` to confirm will
 kill everything.