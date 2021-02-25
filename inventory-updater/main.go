/*

  websocket gorilla echo server

*/

package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

var produces = [...]string{"Eggplant", "Cabbage", "Broccoli", "Lettuce", "Bell Pepper", "Spinach"}
var locations = [...]string{"ABC Farms", "Fiddler Agri Business", "J and J Veggie Farm", "Acme Farm", "Winslow Cooperative"}

var addr = flag.String("addr", "localhost:8080", "http service address")
var templates map[string]*template.Template
var upgrader = websocket.Upgrader{} // use default options
var ws *websocket.Conn
var sentData string

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
		log.Fatal(err)
	}
	s.nc = nc

	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	s.sc = sc
	return s, nil
}

type Item struct {
	Name          string         `json:"name"`
	Qty           int            `json:"qty"`
	ItemLocations []ItemLocation `json:"itemLocations"`
}

type ItemLocation struct {
	LocationName string `json:"locationName"`
	ItemName     string `json:"itemName"`
	Qty          int    `json:"qty"`
}
type Harvest struct {
	ReportDate string `json:"reportDate"`
	Items      []Item `json:"items"`
}

func createItem(name string) Item {
	var itemLocations []ItemLocation
	total := 0
	for _, loc := range locations {
		itemLocation := ItemLocation{
			LocationName: loc,
			ItemName:     name,
			Qty:          rand.Intn(300) + 50,
		}

		total += itemLocation.Qty

		itemLocations = append(itemLocations, itemLocation)
	}

	return Item{
		Name:          name,
		Qty:           total,
		ItemLocations: itemLocations,
	}
}

func (s *server) publishHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	layoutUS := "January 2, 2006"
	t := time.Now()
	//log.Println(t.Format(layoutUS))

	var harvest Harvest
	for _, v := range produces {
		item := createItem(v)
		harvest.Items = append(harvest.Items, item) // Item{Name: v, Qty: rand.Intn(400) + 100})
	}
	harvest.ReportDate = t.Format(layoutUS)
	// log.Println("---- harvests\n", harvest)

	b, err := json.Marshal(harvest)
	if err != nil {
		log.Fatalf("Error on marshalling data: %v\n", err)
	}

	//log.Println("---- after marshal", string(b))

	subj := "inventory.count.update"

	err = s.sc.Publish(subj, b) //[]byte(mssg))
	if err != nil {
		log.Fatalf("Error during publish: %v\n", err)
	}

	//log.Printf("....Published via stream [%s] : '%s'\n", subj, string(b))

	sentData = string(b)

	http.Redirect(w, r, "/", 302)
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
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, "The template does not exist.", http.StatusInternalServerError)
	}

	var v = struct {
		Host string
	}{wshost}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := tmpl.Execute(w, v)
	if err != nil {
		log.Println("Error on template execute", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func home(w http.ResponseWriter, r *http.Request) {
	// log.Println("+++ home Host:", r.Host)
	renderTemplate(w, "index", "base", "ws://"+r.Host+"/data")
}

func postWebData(w http.ResponseWriter, r *http.Request) {
	var err error

	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error on websocket upgrade:", err)
		return
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte(sentData)) // send back the message
	if err != nil {
		log.Println("Error on socket write:", err)
	}

	// log.Println(".... websocket initialize")

}

func main() {

	log.SetFlags(0)
	var (
		clusterID string
		clientID  string
		URL       string
	)

	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clientID, "id", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&clientID, "clientid", "stan-pub", "The NATS Streaming client ID to connect with")

	flag.Usage = usage
	flag.Parse()

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Publisher")}

	serv, err := NewServer(URL, clusterID, clientID, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer serv.sc.Close()
	defer serv.nc.Close()
	defer ws.Close()

	r := mux.NewRouter().StrictSlash(false)
	fs := http.FileServer(http.Dir("public"))
	r.Handle("/public/", fs)
	r.HandleFunc("/", home)
	r.HandleFunc("/data", postWebData)
	r.HandleFunc("/send", serv.publishHandler)
	server := &http.Server{
		Addr:         *addr,
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Listening... port", server.Addr)
	log.Fatal(server.ListenAndServe())

}
