package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Bend struct {
	// Get the method, the payload and the type
	Methods  []string
	Endpoint string
	SQLdb    string
	// If Protected, check bend key vs bend-auth: GUID (else, 403)
	// Fetch the config, the env.txt and the sql.db
	TokenRequired bool
	Runtime       string
}

type Project struct {
	Name    string
	ShortID string
	Token   string
	Bends   []Bend
}

// Look for string in slice
// replace with exp/slices
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func main() {

	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

	router := mux.NewRouter().StrictSlash(true)

	projects := []Project{
		{
			Name:    "Demo",
			ShortID: "g6w0qh",
			Token:   "legit-token",
			Bends: []Bend{
				{
					Endpoint:      "foo",
					Methods:       []string{http.MethodGet, http.MethodPost},
					TokenRequired: true,
					Runtime:       "go",
				},
			},
		},
	}

	// # Project Creator
	// Name it, create it as a shortUUID
	// i.e. app.bendly.io/g6woqh/
	//      app.bendly.io/g6woqh/edit/webhookcatcher
	// Generate the docker image and save the bits for the project

	// # Bend Creator
	// Specify name of endpoint (max length 50)
	// i.e. app.bendly.io/g6woqh/webhookcatcher
	// Specify Require auth [X]
	//  // If yeah, show their auth key
	// Specify methods [POST, GET]
	// Specify runtime [Go]
	// Specify env [env.txt]
	// Specify if logs enabled [X]
	//   // Basically track in, exec logs, out and responses
	// Provide database [X]
	//   // Provide sqlite.db or start blank

	// # Bend Editor
	// i.e. curl  -H "Authorization: legit-token" localhost:8080/edit/g6w0qh/endpoint -v
	router.HandleFunc("/edit/{bend}/{endpoint}", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		// TODO replace with DB
		for i := range projects {
			if auth == projects[i].Token && projects[i].ShortID == mux.Vars(r)["bend"] {
				log.Println("Securely found bend", projects[i].ShortID)
				for j := range projects[i].Bends {
					if projects[i].Bends[j].Endpoint == mux.Vars(r)["endpoint"] {
						fmt.Fprintln(w, "Editing Project", mux.Vars(r)["bend"], "for endpoint", mux.Vars(r)["endpoint"])
						return
					}
				}
			}
		}

		w.WriteHeader(http.StatusForbidden)

	}).Methods(http.MethodPost, http.MethodPost, http.MethodDelete, http.MethodGet)

	// Bend handler
	router.HandleFunc("/{project}/{bend}", func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		project := mux.Vars(r)["project"]
		bend := mux.Vars(r)["bend"]

		// TODO replace with DB lookup
		for i := range projects {
			if projects[i].ShortID == project {
				for j := range projects[i].Bends {
					if projects[i].Bends[j].Endpoint == bend {
						bend := projects[i].Bends[j]
						if auth != projects[i].Token && bend.TokenRequired {
							w.WriteHeader(http.StatusForbidden)
							return
						}

						// Don't pass through bad methods
						if !contains(projects[i].Bends[j].Methods, r.Method) {
							w.WriteHeader(http.StatusMethodNotAllowed)
							return
						}

						log.Println("Retrieving", bend.Endpoint, "from project", project)

						// Get the method, the payload and the type
						// Find the project (else, 404)
						// Find the Bend (else 404)
						// If Protected, check bend key vs Authorization: GUID (else, 403)
						// Fetch the config, the env.txt and the sql.db
						// Copy to /tmp/instance.whatever folder
						location, _ := os.MkdirTemp("", fmt.Sprintf("%s-tmp-*", project))
						log.Println("Created folder", location)
						// defer os.RemoveAll(location)

						// Flatten params
						flatparams := ""
						for k, v := range r.URL.Query() {
							flatparams = flatparams + k + "=" + strings.Join(v, ",") + "\n"

						}

						// Flatten headers
						headers := ""
						for k, v := range r.Header {
							headers = headers + k + "=" + strings.Join(v, ",") + "\n"
						}

						// Get payload
						payload, _ := io.ReadAll(r.Body)

						// refactor me out - check error code
						func(location, method, payload, headers, params string, bend Bend) error {
							log.Println("PopulateWorkspace()", location, bend, method, params)

							// Lookup bend in database and grab the env file or stub one out
							scriptFile, err := os.Create(filepath.Join(location, "main."+bend.Runtime))
							if err != nil {
								log.Println("os.Create err=", err)
								return fmt.Errorf("Whoopsy")
							}
							defer scriptFile.Close()
							io.WriteString(scriptFile, godemo) // TODO MAKE THIS HELLA LESS GROSS

							// Lookup bend in database and grab the env file or stub one out
							envFile, err := os.Create(filepath.Join(location, "env"))
							if err != nil {
								log.Println("os.Create err=", err)
								return fmt.Errorf("Whoopsy")
							}
							defer envFile.Close()

							// Write the method
							// Lookup bend in database and grab the env file or stub one out
							payloadFile, err := os.Create(filepath.Join(location, method))
							if err != nil {
								log.Println("os.Create err=", err)
								return fmt.Errorf("Whoopsy")
							}
							defer payloadFile.Close()
							io.WriteString(payloadFile, payload)

							// Write the headers.in
							headersFile, err := os.Create(filepath.Join(location, "headers.in"))
							if err != nil {
								log.Println("os.Create err=", err)
								return fmt.Errorf("Whoopsy")
							}
							defer headersFile.Close()
							io.WriteString(headersFile, headers)

							// Write the params
							paramsFile, err := os.Create(filepath.Join(location, "params"))
							if err != nil {
								log.Println("os.Create err=", err)
								return fmt.Errorf("Whoopsy")
							}
							defer paramsFile.Close()
							io.WriteString(paramsFile, params)

							return nil
						}(location, r.Method, string(payload), headers, flatparams, bend)

						// Execute docker and mount /tmp/instance.whatever and run script
						// TODO move to use exec so we can setup a kill after 10 seconds
						cmdStr := fmt.Sprintf("docker run -v %s:/bridge --env-file=%s --rm --pull=never  %s/%s:latest", location, filepath.Join(location, "env"), project, bend.Endpoint)
						_, _ = exec.Command("/bin/sh", "-c", cmdStr).Output()

						// Check hash of sqlite.db to see if modified - if yes, merge it back into the root
						//  // all instances of things in /g6woqh/sqlout.db -> into main.db
						// https://stackoverflow.com/questions/80801/how-can-i-merge-many-sqlite-databases

						// On completion, scan for response.* in /tmp/instance.whatever
						err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								fmt.Println(err)
								return err
							}
							// Take the first one that isn't a folder
							if strings.Contains(path, "response.") && !info.IsDir() {
								f, err := os.Open(path)
								if err != nil {
									return fmt.Errorf("os.Open %s %s", err, path)

								}
								w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(info.Name())))
								defer f.Close()
								io.Copy(w, f)
								return nil
							}
							return nil
						})
						if err != nil {
							// If no response, return error code from container 200 OK or 500 Internal Server Error
							w.WriteHeader(http.StatusInternalServerError)
						}
						return
					}
				}
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions)

	// Create the HTML Server
	address := "0.0.0.0"
	port := "8080"
	log.Printf("Starting server on http://%s:%s", address, port)
	server := http.Server{
		Addr:           fmt.Sprintf("%s:%s", address, port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   0 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("API server failed to start with error: %v\n", err)
	}
	log.Println("API server stopped")
}
