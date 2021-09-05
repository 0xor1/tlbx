tlbx [![Coverage Status](https://coveralls.io/repos/github/0xor1/tlbx/badge.svg)](https://coveralls.io/github/0xor1/tlbx)
====

## fmt & test

```
./bin/pre
```
to gen, fmt and test everything.

```
./bin/pre nuke
```
to clear out all docker containers and rebuild before gen, fmt and test.
(useful if you have a sql schema change and want to recreate your sql db containers).

## structure

* `/pkg` - reusable packages
* `/bin` - util scripts to run/build/test
* `/cmd` - executable go programs
* `/sql` - sql schemas
* `/docker` - docker-compose files for dev environment setup

## web apps

tlbx is predominantly for making web apps that follow a similar pattern, to run these apps, simply run
`./bin/run <app_name>` e.g. `./bin/run todo`. This script makes use of tmux to run all the aspects of an
app in one terminal screen so you can see everything going on in one place, if you prefer you can simply run the
commands in the `./bin/run` script manually. If you do install tmux, to kill the development services tmux
cmd `Ctrl+b &` then `y` to confirm will kill everything.

* [todo](https://github.com/0xor1/tlbx/tree/develop/cmd/todo) - typical todo list demo app, the most simple/minimal demo fo tlbx framework
* [games](https://github.com/0xor1/tlbx/tree/develop/cmd/games) - a multiplayer turn based game site
* [trees](https://github.com/0xor1/tlbx/tree/develop/cmd/trees) - a project management app where tasks are stored in a tree structure