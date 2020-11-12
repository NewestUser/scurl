package main

import (
	"flag"
	"fmt"
	"github.com/newestuser/scurl/lib"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const Version = "0.6"

func main() {
	fs := flag.NewFlagSet("scurl", flag.ExitOnError)

	version := fs.Bool("version", false, "Print version and exit")

	opts := &reqOpts{
		headers: headers{make([]string, 0)},
		method:  methodFlag{},
		rate:    rateFlag{scurl.DefaultRate},
		verbose: false,
		form:    multipartForm{map[string]string{}},
	}

	fs.IntVar(&opts.fanOut, "fo", 1, "Fan out factor is the number of clients to spawn")
	fs.Var(&opts.rate, "rate", "Rate of the requests to be send by the client (i.e. 50/1s)")
	fs.DurationVar(&opts.duration, "duration", 0, "Duration of stress [0 = forever] (i.e. 1m) (default 0)")
	fs.Var(&opts.method, "X", "HTTP method to use (default GET)")
	fs.Var(&opts.headers, "H", "HTTP header to add")
	fs.StringVar(&opts.body, "d", "", "HTTP body to transport")
	fs.Var(&opts.form, "F", "Add form-data in the format [key=value] (Content-Type is set to multipart/form-data)")
	fs.BoolVar(&opts.verbose, "verbose", false, "Verbose logging")

	fs.Usage = func() {
		fmt.Println("Usage: scurl [global flags] '<url>'")
		fmt.Printf("\nglobal flags:\n")
		fs.PrintDefaults()
		fmt.Print(example)
		return
	}

	cmdArgs := os.Args[1:]
	if err := fs.Parse(cmdArgs); err != nil {
		log.Fatal(err)
	}

	if *version {
		fmt.Printf("Version: %s\n", Version)
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
	bodyOption, err := opts.bodyOption()
	if err != nil {
		return err
	}

	request, e := scurl.NewTarget(target,
		scurl.MethodOption(opts.method.verb),
		bodyOption,
		scurl.HeaderOption(opts.headers.headers...),
	)

	if e != nil {
		return e
	}

	client := scurl.NewConcurrentClient(
		scurl.FanOutOpt(opts.fanOut),
		scurl.RateOpt(opts.rate.val),
		scurl.DurationOpt(opts.duration),
		scurl.VerboseOpt(opts.verbose),
	)

	res := client.DoReq(request)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

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

			r.ReadAndDiscard()
			concurrentResp.Add(r)
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
		fmt.Println("Total bytes:", resp.TotalBites())

		statMap := resp.StatusMap()
		for status, resps := range statMap {
			fmt.Printf("\tStatus %d: %d responses\n", status, len(resps))
		}
	}
}

const example = `
example:
	scurl -rate 50/1s -X POST -H 'Content-Type: application/json' -d '{"key":"val"}' 'http://localhost:8080'
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

type multipartForm struct {
	values map[string]string
}

func (f *multipartForm) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) == 2 {
		f.values[parts[0]] = parts[1]
		return nil
	} else if len(parts) == 1 {
		f.values[parts[0]] = ""
		return nil
	}

	return fmt.Errorf("form data '%s' has a wrong format", value)
}

func (f multipartForm) String() string {
	return fmt.Sprintf("%s", f.values)
}

type reqOpts struct {
	verbose  bool
	fanOut   int
	rate     rateFlag
	duration time.Duration

	method  methodFlag
	headers headers
	body    string
	form    multipartForm
}

func (o reqOpts) bodyOption() (scurl.ReqOption, error) {
	if len(o.body) != 0 && len(o.form.values) != 0 {
		return nil, fmt.Errorf("cannot provide both HTTP body '-d' and form-urlencoded data '-F'")
	}

	if len(o.body) != 0 {
		return scurl.StringBodyOption(o.body), nil
	}

	return scurl.MultipartFormBodyOption(o.form.values), nil
}

type methodFlag struct {
	verb string
}

func (m *methodFlag) String() string {
	return m.verb
}

// Set implements the flag.Value interface for HTTP methods.
func (m *methodFlag) Set(val string) error {
	if len(val) == 0 {
		m.verb = scurl.DefaultMethod
		return nil
	}

	allowed := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	contained := false
	for _, v := range allowed {
		if v == val {
			contained = true
			break
		}
	}

	if !contained {
		return fmt.Errorf("method '%s' is not supported, supported methods are %s", val, allowed)
	}

	m.verb = val
	return nil
}

type rateFlag struct {
	val *scurl.Rate
}

func (r *rateFlag) String() string {
	if r.val == nil {
		return ""
	}

	return r.val.String()
}

// Set implements the flag.Value interface for request rate limiting.
func (r *rateFlag) Set(val string) error {
	parts := strings.Split(val, "/")

	if len(parts) == 0 {
		return fmt.Errorf(`-rate format %q does not match the "freq/duration" format (i.e. 50/1s)`, val)
	}

	if len(parts) == 1 {
		parts = append(parts, `1s`)
	}

	freq, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf(`-rate format %s does not match the "freq/duration" format (i.e. 50/1s)`, parts)
	}

	switch parts[1] {
	case "ns", "µs", "ms", "s", "m", "h":
		parts[1] = "1" + parts[1]
	}

	duration, err := time.ParseDuration(parts[1])
	if err != nil {
		return fmt.Errorf("-rate format %s does not match the \"freq/duration\" format (i.e. 50/1s), "+
			`possible time units ["ns", "µs", "ms", "s", "m", "h"]`, val)
	}

	r.val.Freq = freq
	r.val.Per = duration

	if r.val.IsZero() {
		return fmt.Errorf("-rate value cannot be zero, both freq and duration need to be > 0 (i.e. 50/1s)")
	}

	return nil
}
