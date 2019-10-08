# deployroute

`deployroute` will deploy a route for the ColorApp demo as a single or rolling update.

## Usage

```
USAGE: deployroute [OPTIONS] INITIAL-VERSION FINAL-VERSION [STEP]

Required:
  Environment variable APP must be set to the URL for the app.

Commands:
  INITIAL-VERSION   Initial color (i.e., blue|green|red).
  FINAL-VERSION     Final color (i.e., blue|green|red).
  STEP              Optional percent, such as 10, for rolling update increment.
                    If STEP is not set, the deployment is performed in 1 stage.

Options:
  --count,-c        The number of tests performed at each stage (default=0).
  --threshold,-t    The threshold as a percent of acceptable errors (default=0).
                    Errors are HTTP 400 and greater response status codes.
                    The percent applies against the running total number of tests.
  --help,-h         Display this help.


Test options and arguments are specific to the Color App:
- The only valid deployment versions correspond to `blue`, `green`, and `red`.
- The only test that is performed is a HTTP GET to the color endpoint (i.e.,
  $APP/color).
```

## Building deployroute

You must [install Go](https://golang.org/dl/) to build the tool for your system.

You can build `deployroute` in this directory with the command:

    $ go build .

Alternatively, if you have set up a [valid workspace directory](https://golang.org/doc/install#testing)
for Go, you can also install `deployroute`. If you have a standard workspace
(i.e., `$HOME/go`), then installing the tool will create an executable at
`$HOME/go/bin/deployroute`.

    $ go install

Ensure that `$HOME/go/bin` is included in your shell path.

