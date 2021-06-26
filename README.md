# label-operator

```sh
make run
```

## operator-sdk create controllers steps

1. operator-sdk init --domain=lowang.io --repo=github.com/lowang-bh/label-operator
2. operator-sdk create api --group=core --version=v1 --kind=Pod --controller=true --resource=false
3. write logic
