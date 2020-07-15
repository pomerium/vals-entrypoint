# About

In some container environments, secret sources are not always readily available or secure.  Serverless options such as GCP's Cloud Run require your application to directly
interface with Secret Manager or another source.  vals-entrypoint is intended to serve as an exec wrapper for your applications to launch in environments where secret material is to be retrieved by using ambient credentials such as a GCE service account.

vals-entrypoint supports substituting secrets into environment variables, as well as writing secrets to the filesystem.  To do this, it relies on [vals](https://github.com/variantdev/vals) to provide integrations with major cloud providers and on-prem secret stores.

- All environment variables containing vars secretrefs will be evaluated
- `VARS_FILES` environment variable can be set to a list of file/secretref pairs to write the corresponding secretref to a file
- `--vars-files` flag can be set to a similar list of file/secretref pairs to be written out

# Examples

Given an existing secret named `mysecret`:
```yaml
mykey: myvalue
```

Load entire secret into variable FOO
```shell
FOO="ref+gcpsecrets://myproject/mysecret" vars-entrypoint exec -- /bin/sh -c 'echo ${FOO}'
mykey: myvalue
```

Load subkey into variable FOO
```shell
FOO="ref+gcpsecrets://myproject/mysecret#/mykey" vars-entrypoint exec -- /bin/sh -c 'echo ${FOO}'
myvalue
```

Write entire secret into config using `VARS_FILES`
```shell
VARS_FILES="/tmp/config.yaml:ref+gcpsecrets://myproject/mysecret" vars-entrypoint exec cat /tmp/config.yaml
mykey: myvalue
```

Write entire secret into file using `--vars-files`
```shell
vars-entrypoint --vars-files="/tmp/config.yaml:ref+gcpsecrets://myproject/mysecret" exec cat /tmp/config.yaml
mykey: myvalue
```

# Dockerfile

An easy way to create a container image with vals-entrypoint is to build an image using `pomerium/vals-entrypoint` as a named layer and copy the binary in:

```Dockerfile
FROM pomerium/vals-entrypoint as base

FROM pomerium/pomerium:latest
COPY --from=base /bin/vals-entrypoint /bin/vals-entrypoint

ENTRYPOINT ["/bin/vals-entrypoint"]
CMD ["/bin/pomerium","-config","/pomerium/config.yaml"]
```
