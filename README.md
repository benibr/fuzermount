# fuzermount

This is a wrapper tool for `/usr/bin/fusermount3` to control who is calling fusermount with what options.
It's meant to take fusermounts place with SUID and root ownership and if the original call is allowed
it calls the actual fusermount binary to do the work.
This way it's possible to remove SUID/GUID from fusermount completely.

## Usage

Fuzermount has 3 levels of enforcing specific mount parameters.
When `mode=fallthrough` is set, fuzermount drops all privileges coming from the SUID bit
and passes all arguments to the target. This mode is useful for logging.
When `mode=relaxed` is set, mounting is denied only if explicitly forbidden mount options are found.
Otherwise it behaves like `fallthrough` mode.
When `mode=strict` is set, the arguments are checked for forbidden and mandatory options,
then it's checked if the calling binary is allowed and finally the (u)mount arguments are rebuild
to drop everything unknown.

See [./fuzermount.yaml](./fuzermount.yaml) for an example configuration.

### Compiling

```
# build binary
make

# build rpm package
make rpm

# build container with RPM installed and mocked applications
make containerbuild
```

## Build dependencies

* go: to build
* nfpm: for rpm building
* podman: for containerbuilds and testing
