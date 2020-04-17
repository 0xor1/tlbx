wtf
===

## fmt & test

```
./bin/pre
```

## structure

* `/pkg` - reusable packages
* `/bin` - util scripts to run/build/test
* `/cmd` - executable go programs
* `/common` - common clientside ui commponents and sql db schemas

## web apps

wtf is predominantly for making web apps that follow a similar pattern, to run these apps tmux, go,
 node/npm and docker/docker-compose must be installed then simply `./bin/run <app_name>` e.g.
 `./bin/run todo` and open a browser tab to `localhost:8081`. to kill the development services tmux
 cmd `Ctrl+b &` then `y` to confirm will kill everything.