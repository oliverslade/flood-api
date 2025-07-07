# DIY Flood API

## The challenge

You have been provided with an OpenAPI contract, an SQLite database, and an executable that runs a series of tests against an API.

Your task is to build an API in whatever language and framework you like to match the API contract, and make all of the tests pass.

**Using a language or framework you're less familiar with will earn you bonus cred!**

You should modify the structure of the database, but every call to your API should read from the database (so no caching it in your solution). You may also migrate the data to another database solution if you want.

Without any modification to the database, the tests will run in around 40 seconds, so some performance analysis and optimisation is recommended!

### The data

The data is river level and rainfall data from [Defra's flood monitoring API](https://environment.data.gov.uk/flood-monitoring/doc/reference) that has been collected every day for over two years.

## Testing your solution

I don't have an Apple developer account, so haven't been able to create a code-signed version of the testing program.

In order to run the executable you'll have to first un-quarantine the binary

```sh
xattr -d com.apple.quarantine ./FloodApiTests
```

> If you don't want to run the binary, feel free to write your own tests to check the correctness of your solution.

By default the executable runs all of its tests against `http://localhost:9001`, but that can be changed by adding an optional parameter.

```sh
# to run with the default URL
# localhost port 9001
./FloodApiTests

# to target an alternative URL
# e.g. localhost port 3000
./FloodApiTests http://localhost:3000
```

> Note that the protocol (`http`) is required, and all of your tests will probably fail without it.

A successful run will look something like this:

```sh
$ ./FloodApiTests
Tests starting for http://localhost:9001/
Finished test suite in 935ms, 164/164 tests passed
```

The test suite has been compiled for `osx-arm64` so if you're using an older Intel Macbook please let me know and I'll try to compile a new test suite for you. My assumption is that by now every consultant at OC has an M1 or newer Macbook.
