/*

  websocket gorilla echo server

*/

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

var templates map[string]*template.Template

var upgrader = websocket.Upgrader{} // use default options
var ws *websocket.Conn

type server struct {
	nc *nats.Conn
	sc stan.Conn
}

var usageStr = `
Usage: stan-pub [options] <subject> <message>

Options:
	-s,  --server   <url>            NATS Streaming server URL(s)
	-c,  --cluster  <cluster name>   NATS Streaming cluster name
	-id, --clientid <client ID>      NATS Streaming client ID
	-a,  --addr     <server address> This server address
`

// NOTE: Use tls scheme for TLS, e.g. stan-pub -s tls://demo.nats.io:4443 foo hello
func usage() {
	log.Printf("%s\n", usageStr)
	os.Exit(0)
}

func NewServer(URL, clusterID, clientID string, opts []nats.Option) (server, error) {

	var s server

	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Println("Error on connecting to nats server url", URL, " Message:", err)
		log.Fatal(err)
	}
	s.nc = nc

	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Println("Error on connecting to stan server url", URL, " Message:", err)
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	s.sc = sc
	return s, nil
}

func postWebData(w http.ResponseWriter, r *http.Request) {
	var err error

	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error on websocket upgrade:", err)
		return
	}

	strMessage := ""
	err = ws.WriteMessage(websocket.TextMessage, []byte(strMessage)) // send back the message
	if err != nil {
		log.Println("Error on socket write:", err)
	}

	//log.Println(".... websocket initialize")

}

//Compile view templates
func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	templates["index"] = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
}

//Render templates for the given name, template definition and data object
func renderTemplate(w http.ResponseWriter, name string, template string, wshost string) {

	//log.Println("....template name:", name, "   host", wshost)

	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "The template does not exist.", http.StatusInternalServerError)
	}
	var v = struct {
		Host string
	}{wshost}

	//w.Header().Set("Content-Type", "text/html; charset=utf-8; application/javascript;charset=utf-8")
	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	// w.Header().Set("X-Content-Type-Options", "nosniff")
	err := tmpl.Execute(w, v)
	if err != nil {
		log.Println("Error on template execute", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	log.Println("Host:", r.Host)
	renderTemplate(w, "index", "base", "ws://"+r.Host+"/data")
}

func main() {

	log.SetFlags(0)

	var (
		clusterID string
		clientID  string
		URL       string
		addr      string
	)

	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clientID, "id", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&clientID, "clientid", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&addr, "a", "localhost:7575", "This server address")
	flag.StringVar(&addr, "addr", "localhost:7575", "This server address")

	flag.Usage = usage
	flag.Parse()

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Publisher")}

	serv, err := NewServer(URL, clusterID, clientID, opts)
	if err != nil {
		log.Println("Error on creating new stan server", err)
		log.Fatal(err)
	}
	defer serv.sc.Close()
	defer serv.nc.Close()
	defer ws.Close()

	r := mux.NewRouter().StrictSlash(false)

	fs := http.FileServer(http.Dir("public"))
	r.Handle("/public/", fs)
	r.HandleFunc("/data", postWebData)
	r.HandleFunc("/", home)

	subj := "inventory.count.update"

	mcb := func(msg *stan.Msg) {
		// log.Println(".....data received:\n", string(msg.Data))

		if ws != nil {
			err = ws.WriteMessage(websocket.TextMessage, msg.Data) // send back the message
			if err != nil {
				log.Println("Error on write to socket:", err)
			}
		}
	}

	qgroup := ""
	durable := ""
	startOpt := stan.StartAtSequence(1)
	_, err = serv.sc.QueueSubscribe(subj, qgroup, mcb, startOpt, stan.DurableName(durable))
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Listening... port", server.Addr)
	log.Fatal(server.ListenAndServe())
}
