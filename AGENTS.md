# notes for openai's codex agent

please set up this project properly, while internet
access is still available, by executing `codex_setup.sh`.

after working, make sure tests pass by running `go test`.

ensure proper formatting with `gofmt`, either on individual files,
or perhaps like this for the whole project:

```
find . -name "*.go" -exec gofmt -w {} \;
```
