package main

import (
	"fmt"
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
//	"github.com/golang/glog"
	"context"
)


type myServerHandler struct {

}

var (
	tlscert , tlskey string
)

func getEnv(key , fallback string) string {
	value , exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func main() {
	
	// chaeck the Environment variable for User Define Placements
	certpem := getEnv("CERT_FILE", "/etc/certs/cert.pem")
	keypem := getEnv("KEY_FILE", "/etc/certs/key.pem")
	port := getEnv("PORT", "8443")

	flag.StringVar(&tlscert, "tlsCertFile", certpem , "The File Contains the X509 Certificate for HTTPS")
	flag.StringVar(&tlskey, "tlsKeyFile", keypem , "The File Contains the X509 Private key")

	flag.Parse()

	certs , err := tls.LoadX509KeyPair(tlscert, tlskey)

	if err != nil {
//		glog.Errorf("Failed to load Certificate/Key Pair: %v", err)
        fmt.Fprintf(os.Stderr, "Failed to load Certificate/Key Pair: %v", err);
	}

	// Setting the HTTP Server with TLS (HTTPS)
	server := &http.Server {
		Addr: fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}},
	}

	// Setting 2 variable which are defined by an empty struct for each of the function depending on the URL path 
	// the http request is calling 
	// in our example we have 2 paths , one for the mutate and one for validate

	mr := myServerHandler{}
	gs := myServerHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", mr.mutserve)
	mux.HandleFunc("/validate", gs.valserve)
	server.Handler = mux


	// Starting a new channel to start the Server with TLS configuration we provided when we defined the server 
	// variable 

	go func() {
		if err := server.ListenAndServeTLS("",""); err != nil {
//			glog.Errorf("Failed to Listen and Serve Web Hook Server: %v", err)
			fmt.Fprintf(os.Stderr, "Failed to Listen and Serve Web Hook Server: %v", err);
		}
	}()

//	glog.Infof("The Server Is running on Port : %s ", port)
	fmt.Fprintf(os.Stdout, "The Server Is running on Port : %s \n" , port)

	// Next we are going to setup the single handling for our HTTP server by sending the right signals to the channel

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

//	glog.Info("Get Shutdown signal , sutting down the webhook Server gracefully...")
    fmt.Fprintf(os.Stdout, "Get Shutdown signal , sutting down the webhook Server gracefully...\n")
	server.Shutdown(context.Background())

}