# Communications

To run a local simulation you will need to:

```
$ make
$ build/coordinator --noopAuthEnabled
$ build/server --noopAuthEnabled --authMethod=noop
$ build/bots -n 10 --subscribe
```

You can optionally start an extra bot to print stats

```
$ build/bots -n 1 --subscribe --trackStats
```
