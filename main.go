package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
	"strings"
	"github.com/newestuser/scurl/lib"
	"os/signal"
	"time"
	"log"
	"runtime"
)

const Version = "0.2"

func main() {
	fs := flag.NewFlagSet("scurl", flag.ExitOnError)

	version := fs.Bool("version", false, "Print version and exit")

	opts := &reqOpts{
		headers: headers{make([]string, 0)},
		method:  method{},
	}
	fs.IntVar(&opts.fanOut, "fo", 10, "Number of requests to send concurrently")
	fs.Var(&opts.method, "X", "HTTP method to use")
	fs.Var(&opts.headers, "H", "HTTP header to add")
	fs.StringVar(&opts.body, "d", "", "HTTP body to transport")

	fs.Usage = func() {
		fmt.Println("Usage: scurl [global flags] <url>")
		fmt.Printf("\nglobal flags:\n")
		fs.PrintDefaults()
		fmt.Println(example)
		return
	}
	cmdArgs := os.Args[1:]
	if err := fs.Parse(cmdArgs); err != nil {
		log.Fatal(err)
	}

	if *version {
		fmt.Printf("Version: %s", Version)
		return
	}

	if len(fs.Args()) != 1 {
		fs.Usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	if e := stress(fs.Args()[0], opts); e != nil {
		log.Fatal(e.Error())
	}

}

func stress(target string, opts *reqOpts) error {

	request, e := scurl.NewRequest(target,
		scurl.MethodOption(opts.method.verb),
		scurl.BodyOption(opts.body),
		scurl.HeaderOption(opts.headers.headers...),
	)

	if e != nil {
		return e
	}

	client := scurl.NewConcurrentClient(opts.fanOut)

	res := client.DoReq(request)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	concurrentResp := &scurl.MultiResponse{StartTime: time.Now()}
	for {
		select {
		case <-sig:
			client.Stop()
			printResult(concurrentResp)
			return nil
		case r, ok := <-res:

			if !ok {
				printResult(concurrentResp)
				return nil
			}

			concurrentResp.Add(r)
			r.Body.Close()
			stopWhenAllRespReceived(client, concurrentResp, opts.fanOut)
		}
	}

}

func printResult(resp *scurl.MultiResponse) {
	fmt.Println("Trips:", resp.Trips)
	if !resp.Empty() {
		fmt.Println("Total time:", resp.TotalTime())
		fmt.Println("Avr time:", resp.AvrTime())
		fmt.Println("Fastest:", resp.Fastest().Time)
		fmt.Println("Slowest:", resp.Slowest().Time)

		statMap := resp.StatusMap()
		for status, resps := range statMap {
			fmt.Printf("\tStatus %d: %d responses\n", status, len(resps))
		}
	}
}

func stopWhenAllRespReceived(client *scurl.ConcurrentClient, allResp *scurl.MultiResponse, fanOut int) {
	if len(allResp.Responses) == fanOut {
		client.Stop()
	}
}

const example = `
example:
	scurl -fo 100 -X POST -H "Content-Type: application/json" -d "{\"key\":\"val\"}" http://localhost:8080
`

// headers are the http header parameters used in each request
// it is defined here to implement the flag.Value interface
// in order to support multiple identical flags for request header
// specification
type headers struct{ headers []string }

func (h headers) String() string {
	return fmt.Sprintf("%s", h.headers)
}

// Set implements the flag.Value interface for a map of HTTP Headers.
func (h *headers) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("header '%s' has a wrong format", value)
	}
	key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if key == "" || val == "" {
		return fmt.Errorf("header '%s' has a wrong format", value)
	}

	h.headers = append(h.headers, value)
	return nil
}

type reqOpts struct {
	fanOut  int
	method  method
	headers headers
	body    string
}

type method struct {
	verb string
}

func (m method) String() string {

	return string(m.verb)
}

// Set implements the flag.Value interface for HTTP methods.
func (m *method) Set(value string) error {
	allowed := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	contained := false
	for _, v := range allowed {
		if v == value {
			contained = true
			break
		}
	}

	if !contained {
		return fmt.Errorf("method '%s' is not supported, supported methods are %s", value, allowed)
	}

	m.verb = value
	return nil
}
