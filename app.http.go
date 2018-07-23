package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/freman/awsrangenf/sns"
	gct "github.com/freman/go-commontypes"
	"github.com/freman/work/bootstrap"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func orError(w http.ResponseWriter, code int, err error) error {
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		http.Error(w, err.Error(), code)
	}
	return err
}

func (a *app) runServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", a.indexHandler())
	if a.box != nil {
		r.PathPrefix("/js").Handler(http.FileServer(a.box.HTTPBox()))
		r.PathPrefix("/css").Handler(http.FileServer(a.box.HTTPBox()))
	}
	r.HandleFunc("/api/v1/config", a.configHandler())
	r.HandleFunc("/api/v1/custom", a.customHandler())
	r.HandleFunc("/api/v1/dashboard", a.dashboardHandler())
	r.HandleFunc("/api/v1/imports", a.importsHandler())
	r.HandleFunc("/hook/{key}", a.hookHandler())

	corsHandler := handlers.CORS(
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodOptions}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"Origin", "Content-Type"}),
	)

	a.httpServer = &http.Server{
		Addr:         a.config.Listen,
		Handler:      corsHandler(r),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()
}

func (a *app) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.box == nil {
			fmt.Fprintln(w, "No box found, you sure you aren't running `npm run serve`")
			return
		}
		indexstr, err := a.box.String("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		index, err := template.New("index").Delims(`<script type=config>`, `</script>`).Parse(indexstr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		index.Execute(w, struct{ Embedded template.HTML }{Embedded: `<script>apiUrl=window.location.protocol + "//" + window.location.host + "/api/v1/";</script>`})
	}
}

func (a *app) hookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.config.Webhook.Enabled {
			http.NotFound(w, r)
		}
		vars := mux.Vars(r)
		if key, found := vars["key"]; !found || key != a.config.Webhook.Key {
			http.NotFound(w, r)
		}

		msgTopic := r.Header.Get("X-Amz-Sns-Topic-Arn")
		if msgTopic != "arn:aws:sns:us-east-1:806199016981:AmazonIpSpaceChanged" {
			http.NotFound(w, r)
		}
		msgType := r.Header.Get("X-Amz-Sns-Message-Type")
		if msgType == "SubscriptionConfirmation" {
			var notificationPayload sns.Payload
			defer r.Body.Close()
			dec := json.NewDecoder(r.Body)
			err := dec.Decode(&notificationPayload)
			if err != nil {
				a.log.Println("Decode subscription confirmation message failed", err)
				return
			}
			verifyErr := notificationPayload.VerifyPayload()
			if verifyErr != nil {
				a.log.Println("Verifying subscription confirmation message failed", err)
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			if _, err := notificationPayload.Subscribe(); err != nil {
				a.log.Println("Verifying subscription confirmation message failed", err)
				http.Error(w, err.Error(), http.StatusForbidden)
			}

			return
		}

		// Out of curiosity... would be cool to see this message
		b, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(b))

		a.update()
		SetRoutes(a)
	}
}

func (a *app) configHandler() http.HandlerFunc {
	type configResponse struct {
		ReadOnly bool
		Config   Config
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch r.Method {
		case http.MethodGet:
			enc.Encode(configResponse{
				ReadOnly: bootstrap.IsWritable(a.configFile) != nil,
				Config:   *a.config,
			})
		case http.MethodPost:
			defer r.Body.Close()
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1e6))
			newcfg := *a.config
			if err := orError(w, http.StatusBadRequest, dec.Decode(&newcfg)); err != nil {
				return
			}
			newcfg.Listen = a.config.Listen

			if err := orError(w, http.StatusInternalServerError, saveConfig(a.configFile, &newcfg)); err != nil {
				return
			}

			a.Reload(&newcfg)

			enc.Encode(configResponse{
				ReadOnly: bootstrap.IsWritable(a.configFile) != nil,
				Config:   *a.config,
			})
		}
	}
}

func (a *app) dashboardHandler() http.HandlerFunc {
	type bootstrapStatus struct {
		Finished bool
		Label    string `json:",omitempty"`
		Error    string `json:",omitempty"`
	}
	type dashboardResponse struct {
		Bootstrap bootstrapStatus
		Cards     map[string]interface{}
		Logs      []string
	}
	type Labelled interface {
		Label() string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch r.Method {
		case http.MethodGet:
			var logs []string
			a.ring.Do(func(r interface{}) {
				if r != nil {
					if s, isa := r.(string); isa && s != "" {
						logs = append(logs, s)
					}
				}
			})
			resp := dashboardResponse{
				Bootstrap: bootstrapStatus{
					Finished: a.run.Finished(),
				},
				Cards: map[string]interface{}{
					"AWS Prefixes":  fmt.Sprintf("%d / %d", len(a.prefixes.Filter(a.selections)), len(a.prefixes.PrefixList)),
					"Custom Routes": len(a.customs),
				},
				Logs: logs,
			}

			if task := a.run.Task(); task != nil {
				if l, isa := task.(Labelled); isa {
					resp.Bootstrap.Label = l.Label()
				}
			}
			if err := a.run.Err(); err != nil {
				resp.Bootstrap.Error = err.Error()
			}

			mod, err := os.Stat(a.store("ip-ranges.json"))
			if err == nil {
				resp.Cards["Last Updated"] = mod.ModTime()
			}
			enc.Encode(resp)
		case http.MethodPost:

		}
	}
}

func (a *app) importsHandler() http.HandlerFunc {
	type Selections struct {
		Count           int
		Total           int
		Filter          []string
		RegionToService map[string][]string `json:",omitempty"`
		ServiceToRegion map[string][]string `json:",omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch r.Method {
		case http.MethodGet:
			enc.Encode(&Selections{
				Filter:          a.selections,
				RegionToService: a.prefixes.RegionToService,
				ServiceToRegion: a.prefixes.ServiceToRegion,
				Count:           len(a.prefixes.Filter(a.selections)),
				Total:           len(a.prefixes.PrefixList),
			})
		case http.MethodPost:
			defer r.Body.Close()
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1e6))
			var tmp []string

			if err := orError(w, http.StatusBadRequest, dec.Decode(&tmp)); err != nil {
				return
			}

			if err := orError(w, http.StatusInternalServerError, saveJSON(a.store("selections.json"), &tmp)); err != nil {
				return
			}

			a.selections = tmp

			if err := orError(w, http.StatusInternalServerError, SetRoutes(a)); err != nil {
				return
			}

			enc.Encode(&Selections{
				Filter:          a.selections,
				RegionToService: a.prefixes.RegionToService,
				ServiceToRegion: a.prefixes.ServiceToRegion,
				Count:           len(a.prefixes.Filter(a.selections)),
				Total:           len(a.prefixes.PrefixList),
			})
		}
	}
}

func (a *app) customHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch r.Method {
		case http.MethodGet:
			enc.Encode(a.customs)
		case http.MethodPost:
			defer r.Body.Close()
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1e6))
			var tmp []*gct.Network

			if err := orError(w, http.StatusBadRequest, dec.Decode(&tmp)); err != nil {
				return
			}

			if err := orError(w, http.StatusInternalServerError, saveJSON(a.store("customs.json"), &tmp)); err != nil {
				return
			}

			a.customs = tmp

			if err := orError(w, http.StatusInternalServerError, SetRoutes(a)); err != nil {
				return
			}

			enc.Encode(&a.customs)
		}
	}
}
