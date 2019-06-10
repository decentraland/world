# Identity

## Building the service

```go
$ make build
```

## Run the service locally

```go
$ ./build/identity 
```

By default this will start the service on the 9091 port.

To sign the tokens by default it will use the keys under `config/identity/defaultKeys`. 

### Key generation

You can generate your own keys

```go
$ ./build/keygen <PATH WHERE TO STORE THE GEN KEY>
```

To make the auth server use the generated set the  `PRIV_KEY_PATH` env variable to point to the generated private key `<PATH WHERE TO STORE THE GEN KEY>/generated.key`

 