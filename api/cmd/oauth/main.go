package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/jinzhu/gorm"
	// Blank for package side effect: loads postgres drivers
	_ "github.com/lib/pq"

	oauth "github.com/sm43/oauth/api"
	api "github.com/sm43/oauth/api/gen/api"
	log "github.com/sm43/oauth/api/gen/log"
)

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		hostF     = flag.String("host", "dev", "Server host (valid values: dev)")
		domainF   = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
		httpPortF = flag.String("http-port", "", "HTTP port (overrides host HTTP port specified in service design)")
		secureF   = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	var (
		logger *log.Logger
	)
	{
		logger = log.New("oauth", false)
	}

	// Database Connection
	var (
		db *gorm.DB
	)
	{
		var err error
		db, err = gorm.Open("postgres", "user=postgres password=postgres dbname=oauth sslmode=disable")
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("Successful Db Connection..!!")
		defer db.Close()

		// Enables log for db queries
		db.LogMode(true)

		// Create user table
		// If already exist then won't change anything
		db.AutoMigrate(oauth.User{})
	}

	// Initialize the services.
	var (
		apiSvc api.Service
	)
	{
		apiSvc = oauth.NewAPI(db, logger)
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var (
		apiEndpoints *api.Endpoints
	)
	{
		apiEndpoints = api.NewEndpoints(apiSvc)
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start the servers and send errors (if any) to the error channel.
	switch *hostF {
	case "dev":
		{
			addr := "http://localhost:8000"
			u, err := url.Parse(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
				os.Exit(1)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h := strings.Split(u.Host, ":")[0]
				u.Host = h + ":" + *httpPortF
			} else if u.Port() == "" {
				u.Host += ":80"
			}
			handleHTTPServer(ctx, u, apiEndpoints, &wg, errc, logger, *dbgF)
		}

	default:
		fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: dev)\n", *hostF)
	}

	// Wait for signal.
	logger.Infof("exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	logger.Info("exited")
}
