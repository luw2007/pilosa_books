//go:generate statik -src=./static

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	pilosa "github.com/pilosa/go-pilosa"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/pflag"

	_ "github.com/luw2007/pilosa_books/statik"
)

const defaultHost = "http://localhost:10101"
const indexName = "users"

var Version = "v0.0.1"

func main() {
	pilosaAddr := pflag.String("pilosa", defaultHost, "host:port for pilosa")
	pflag.Parse()

	server, err := NewServer(*pilosaAddr)
	if err != nil {
		log.Fatalf("getting new server: %v", err)
	}
	//server.testQuery()
	fmt.Printf("users count: %d\n", server.NumUsers)
	server.Serve()
}

type Server struct {
	Address  string
	Router   *mux.Router
	Client   *pilosa.Client
	Index    *pilosa.Index
	Fields   map[string]*pilosa.Field
	NumUsers uint64
}

func NewServer(pilosaAddr string) (*Server, error) {
	server := &Server{
		Address: pilosaAddr,
		Fields:  make(map[string]*pilosa.Field),
	}

	router := mux.NewRouter()
	router.HandleFunc("/", server.HandleStatic).Methods("GET")
	router.HandleFunc("/version", server.HandleVersion).Methods("GET")
	router.HandleFunc("/query", server.HandleQuery).Methods("GET")

	pilosaURI, err := pilosa.NewURIFromAddress(pilosaAddr)
	if err != nil {
		return nil, err
	}
	client, err := pilosa.NewClient(pilosaURI)
	if err != nil {
		return nil, errors.Wrap(err, "getting client")
	}
	index := pilosa.NewIndex(indexName)
	err = client.EnsureIndex(index)
	if err != nil {
		return nil, fmt.Errorf("client.EnsureIndex: %v", err)
	}

	fields := []string{
		"gender",
	}

	for _, fieldName := range fields {
		field := index.Field(fieldName, nil)
		if err != nil {
			return nil, fmt.Errorf("index.Field %v: %v", fieldName, err)
		}
		err = client.EnsureField(field)
		if err != nil {
			return nil, fmt.Errorf("client.EnsureField %v: %v", fieldName, err)
		}

		server.Fields[fieldName] = field
	}

	server.Router = router
	server.Client = client
	server.Index = index
	server.NumUsers = server.getUserCount()
	return server, nil
}

func (s *Server) HandleVersion(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(struct {
		DemoVersion   string `json:"demoversion"`
		PilosaVersion string `json:"pilosaversion"`
	}{
		DemoVersion:   Version,
		PilosaVersion: s.getPilosaVersion(),
	}); err != nil {
		log.Printf("write version response error: %s", err)
	}
}

type versionResponse struct {
	Version string `json:"version"`
}

func (s *Server) getPilosaVersion() string {
	resp, err := http.Get(s.Address + "/version")
	if err != nil {
		log.Printf("problem getting version: %v\n", err)
		return ""
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	version := new(versionResponse)
	json.Unmarshal(body, &version)
	return version.Version
}

func (s *Server) testQuery() error {
	// Send a Row query. PilosaException is returned if execution of the query fails.
	response, err := s.Client.Query(s.Fields["pickup_year"].Row(2013), nil)
	if err != nil {
		return fmt.Errorf("s.Client.Query: %v", err)
	}

	// Get the result
	result := response.Result()
	// Act on the result
	if result != nil {
		bits := result.Row().Columns
		fmt.Printf("Got bits: %v\n", bits)
	}
	return nil
}

func (s *Server) Serve() {
	fmt.Println("running at http://0.0.0.0:8000")
	log.Fatal(http.ListenAndServe(":8000", s.Router))
}

func (s *Server) HandleStatic(w http.ResponseWriter, r *http.Request) {
	log.Println("handling")
	statikFS, err := fs.New()
	if err != nil {
		errorText := "Static assets missing. Run `go generate`"
		http.Error(w, errorText, http.StatusInternalServerError)
		log.Println(errorText)
		return
	}
	http.FileServer(statikFS).ServeHTTP(w, r)
}

type intersectResponse struct {
	Rows     []intersectRow `json:"rows"`
	Seconds  float64        `json:"seconds"`
	NumUsers uint64         `json:"numUsers"`
}

type intersectRow struct {
	Count uint64 `json:"count"`
}

var maxIDMap map[string]uint64 = map[string]uint64{
	"speed_mph":            100,
	"duration_minutes":     100,
	"dist_miles":           40,
	"total_amount_dollars": 100,
}

func (s *Server) HandleQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	q, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf(`{"error": "%s"}`, err))
		return
	}

	response, err := s.Client.Query(s.Index.RawQuery(q), nil)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf(`{"error": "%s"}`, err))
		return
	}
	dif := time.Since(start)

	resp := intersectResponse{}
	resp.NumUsers = s.getUserCount()
	resp.Seconds = float64(dif.Seconds())
	resp.Rows = []intersectRow{intersectRow{uint64(response.Result().Count())}}

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		log.Printf("writing results: %v to responsewriter: %v", resp, err)
	}
}

func (s *Server) getUserCount() uint64 {
	var count uint64 = 0
	for n := 0; n < 3; n++ {
		q := s.Index.Count(s.Fields["gender"].Row(uint64(n)))
		response, _ := s.Client.Query(q, nil)
		count += uint64(response.Result().Count())
	}
	return count
}
