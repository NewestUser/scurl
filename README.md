## Stress cURL

A command line tool similar to cURL that can be used for sending multiple HTTP requests.

## Usage manual
```console 
Usage: scurl [global flags] <url>

global flags:
  -H value
    	HTTP header to add
  -X value
    	HTTP method to use (default GET)
  -d string
    	HTTP body to transport
  -duration duration
    	Duration of stress [0 = forever] (i.e. 1m) (default 0)
  -fo int
    	Fan out factor is the number of clients to spawn (default 1)
  -rate value
    	Rate of the requests to be send by the client (i.e. 50/1s) (default 50/1s)
  -version
    	Print version and exit

example:
	scurl -rate 50/1s -X POST -H "Content-Type: application/json" -d "{\"key\":\"val\"}" http://localhost:8080
```

## Credit
The project is motivated by [Vegeta](https://github.com/tsenart/vegeta).

