## Stress cURL

A command line tool similar to cURL that can be used for sending multiple HTTP requests.

## Usage manual
```console 
Usage: scurl [global flags] <url>

global flags:
  -fo int
        Number of requests to send concurrently (default 10)
  -H value
        HTTP header to add
  -X value
        HTTP method to use
  -d string
        HTTP body to transport
  -version
        Print version and exit

example:
        scurl -fo 100 -X POST -H "Content-Type: application/json" -d "{\"key":\"val\"}" http://localhost:8080
```
