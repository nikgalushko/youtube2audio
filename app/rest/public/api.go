package public

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jetuuuu/youtube2audio/app/config"
	"github.com/jetuuuu/youtube2audio/app/rest/errors"
	"github.com/jetuuuu/youtube2audio/app/rest/interfaces"
	"github.com/jetuuuu/youtube2audio/app/storage"
	"github.com/jetuuuu/youtube2audio/app/utils"
	"github.com/jetuuuu/youtube2audio/app/youtube"

	"github.com/gorilla/feeds"
)

var (
	timings = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "app_method_timing",
			Help: "per method time",
		},
		[]string{"method"},
	)
	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_method_counter",
			Help: "per method count",
		},
		[]string{"method"},
	)
)

type Server struct {
	token     *jwtauth.JWTAuth
	rssToken  *jwtauth.JWTAuth
	cfgReader config.ConfigReader
	s         *storage.Storage
}

func New(c config.ConfigReader, store *storage.Storage) *Server {
	usersToken := jwtauth.New("HS256", []byte("secret"), nil)
	rssToken := jwtauth.New("HS256", []byte("rss_secret"), nil)
	s := Server{token: usersToken, cfgReader: c, s: store, rssToken: rssToken}
	return &s
}

func (s *Server) Run() error {
	log.Printf("[public] Run")

	prometheus.MustRegister(timings)
	prometheus.MustRegister(counter)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {

		r.Use(middleware.Recoverer)
		r.Use(middleware.RequestID)
		r.Use(middleware.Logger)
		r.Use(middleware.RealIP)
		r.Use(middleware.Throttle(10), middleware.Timeout(30*time.Second))
		r.Use(timeTrackMiddleware)

		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(s.token))
			r.Use(jwtauth.Authenticator)

			r.With(linkContext).Get("/audio*", s.getAudioFromLink)
			r.Get("/history", s.history)
			r.Delete("/delete_from_history/{item}", s.clearHistory)
			r.Get("/generate_rss_link", s.generateRssLink)
		})

		r.Group(func(r chi.Router) {
			r.Post("/login", s.login)
			r.Post("/create", s.createUser)
			r.Get("/rss/{rssToken}", s.rss)
		})
	})

	router.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(":8080", router)
	log.Fatal(err)
	return err
}

func (s Server) rss(w http.ResponseWriter, r *http.Request) {
	feed := &feeds.Feed{
		Title:       "jetuuu feed",
		Description: "jetuuu's records",
		Created:     time.Now(),
	}

	t, _ := s.rssToken.Decode(jwtauth.TokenFromHeader(r))
	claims := t.Claims.(jwt.MapClaims)

	u := &storage.User{}
	s.s.Load("users", claims["login"].(string), u)

	for _, h := range u.History {
		var item storage.HistoryItem
		s.s.Load("history", h, &item)
		feed.Items = append(feed.Items, &feeds.Item{
			Title:   item.Title,
			Link:    &feeds.Link{Href: item.Link},
			Created: item.Time,
		})
	}

	rssFeed, err := feed.ToRss()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write([]byte(rssFeed))
	}
}

func (s Server) generateRssLink(w http.ResponseWriter, r *http.Request) {
	t, _ := s.rssToken.Decode(jwtauth.TokenFromHeader(r))
	claims := t.Claims.(jwt.MapClaims)

	_, token, err := s.rssToken.Encode(jwtauth.Claims{"login": claims["login"].(string)})
	if err != nil {
		render.Render(w, r, &errors.Renderer{Status: http.StatusBadRequest, Error: err})
	}

	render.JSON(w, r, interfaces.JSON{"rss_link": "/rss/" + token})
}

func (s Server) getAudioFromLink(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value("url").(*url.URL)
	if !ok {
		render.Render(w, r, errors.InvalidRequest)
		return
	}

	resp, err := http.Get(u.String())
	if err != nil {
		render.Render(w, r, errors.InvalidRequest)
		return
	}

	jobID := utils.Hash(middleware.GetReqID(r.Context()))

	t, _ := s.token.Decode(jwtauth.TokenFromHeader(r))
	claims := t.Claims.(jwt.MapClaims)

	var user storage.User
	s.s.Load("users", claims["login"].(string), &user)

	user.History = append(user.History, jobID)
	s.s.Save("users", claims["login"].(string), &user)

	go s.sendJobToConverter(u, jobID)

	render.JSON(w, r, interfaces.JSON{"code": resp.Status, "jobID": jobID})
}

func (s Server) sendJobToConverter(u *url.URL, id string) {
	var err error
	v, err := youtube.NewFromURL(u)
	if err != nil {
		log.Printf("[%s] [WARN] error in getting info about %s\n", id, u.String())
		return
	}

	defer func() {
		var status string
		if err == nil {
			status = "performed"
		} else {
			status = "fail"
		}
		s.s.Save("history", id, &storage.HistoryItem{Time: time.Now(), Status: status, Title: v.Title})
	}()

	log.Printf("[%s] [INFO] v %s", id, v.Duration)

	cfg := s.cfgReader.Read()
	node := cfg.Converters.Next()
	targetURL := "http://" + node.Adress + "/api/v1/processing"
	log.Printf("[%s] [INFO] send request to %s->%s", id, node.Name, targetURL)

	data, err := json.Marshal(interfaces.JSON{"job_id": id, "link": v.Smallest().URL})
	if err != nil {
		log.Printf("[%s] [WARN] json marshal error %s", id, err.Error())
		return
	}

	log.Printf("[%s] [INFO] json data  %s", id, string(data))
	resp, err := http.Post(targetURL, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("[%s] [WARN] error put job into queue %s", id, err.Error())
		return
	}
	log.Printf("[%s] [INFO] response code %d", id, resp.StatusCode)
}

func (s Server) history(w http.ResponseWriter, r *http.Request) {
	t, _ := s.token.Decode(jwtauth.TokenFromHeader(r))

	log.Printf("[%s] [INFO] info history job %s", middleware.GetReqID(r.Context()))

	claims := t.Claims.(jwt.MapClaims)
	u := &storage.User{}
	s.s.Load("users", claims["login"].(string), u)
	var history []storage.HistoryItem

	for _, h := range u.History {
		var item storage.HistoryItem
		s.s.Load("history", h, &item)
		item.ID = h
		history = append(history, item)
	}

	render.JSON(w, r, interfaces.JSON{"history": history})
}

func (s Server) clearHistory(w http.ResponseWriter, r *http.Request) {
	t, _ := s.token.Decode(jwtauth.TokenFromHeader(r))

	reqID := middleware.GetReqID(r.Context())
	log.Printf("[%s] [INFO] clear history", reqID)

	claims := t.Claims.(jwt.MapClaims)
	u := &storage.User{}
	s.s.Load("users", claims["login"].(string), u)

	item := chi.URLParam(r, "item")
	cfg := s.cfgReader.Read()
	node := cfg.Converters.Next()
	targetURL := "http://" + node.Adress + "/api/v1/delete/" + item
	httpClient := http.Client{Timeout: 5 * time.Second}

	for i, h := range u.History {
		if h == item {
			req, err := http.NewRequest(http.MethodDelete, targetURL, nil)
			if err != nil {
				log.Printf("[%s] [WARN] make request error %s", reqID, err.Error())
				render.Render(w, r, &errors.Renderer{Status: http.StatusInternalServerError, Error: err})
				return
			}

			log.Printf("[%s] [INFO] send request to %s", reqID, targetURL)
			resp, err := httpClient.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				s.s.Delete("history", item)
				u.History = append(u.History[:i], u.History[i+1:]...)
				s.s.Save("users", claims["login"].(string), u)
				break
			} else {
				log.Printf("[%s] [WARN] response code %d", resp.StatusCode)
			}
		}
	}

}

func (s Server) login(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Login string `json:"login"`
		Pass  string `json:"pass"`
	}{}

	defer r.Body.Close()

	if err := render.DecodeJSON(r.Body, &request); err != nil {
		render.Render(w, r, &errors.Renderer{Status: http.StatusBadRequest, Error: err})
		return
	}

	u := &storage.User{}
	err := s.s.Load("users", request.Login, u)
	if err != nil || u.Pass != request.Pass {
		render.Render(w, r, &errors.Renderer{Status: http.StatusUnauthorized})
		return
	}

	now := time.Now()
	_, token, err := s.token.Encode(jwtauth.Claims{"exp": now.Add(30 * time.Minute).Unix(), "login": u.Login})
	if err != nil {
		render.Render(w, r, &errors.Renderer{Status: http.StatusBadRequest, Error: err})
	}

	log.Printf("[%s] [INFO] gave new token for %s \n", middleware.GetReqID(r.Context()), request.Login)
	render.JSON(w, r, interfaces.JSON{"token": token})
}

func (s Server) createUser(w http.ResponseWriter, r *http.Request) {
	u := &storage.User{
		Permissions: storage.Permission{
			TTL:            10 * time.Minute,
			RequestPerHour: 5,
		},
	}

	defer r.Body.Close()

	if err := render.DecodeJSON(r.Body, u); err != nil {
		render.Render(w, r, errors.InvalidRequest)
		return
	}

	if err := s.s.Save("users", u.Login, u); err != nil {
		render.Render(w, r, errors.InvalidRequest)
		return
	}

	render.JSON(w, r, interfaces.JSON{"status": "ok"})
}

func linkContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		link := r.URL.Query().Get("link")
		u, err := url.ParseRequestURI(link)
		if err != nil || u == nil {
			render.Render(w, r, errors.NotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "url", u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func timeTrackMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		timings.WithLabelValues(r.URL.Path).Observe(float64(time.Since(start).Seconds()))
		counter.WithLabelValues(r.URL.Path).Inc()
	}
	return http.HandlerFunc(fn)
}
