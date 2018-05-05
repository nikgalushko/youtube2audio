package public

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"

	"github.com/jetuuuu/youtube2audio/app/config"
	"github.com/jetuuuu/youtube2audio/app/rest/errors"
	"github.com/jetuuuu/youtube2audio/app/rest/interfaces"
	"github.com/jetuuuu/youtube2audio/app/storage"
	"github.com/jetuuuu/youtube2audio/app/utils"
	"github.com/jetuuuu/youtube2audio/app/youtube"
)

type Server struct {
	token     *jwtauth.JWTAuth
	cfgReader config.ConfigReader
	s         *storage.Storage
}

func New(c config.ConfigReader, store *storage.Storage) *Server {
	usersToken := jwtauth.New("HS256", []byte("secret"), nil)
	s := Server{token: usersToken, cfgReader: c, s: store}
	return &s
}

func (s *Server) Run() error {
	log.Printf("public Run")
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.RealIP)
	router.Use(middleware.Throttle(10), middleware.Timeout(30*time.Second))
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(s.token))
			r.Use(jwtauth.Authenticator)

			r.With(linkContext).Get("/audio*", s.getAudioFromLink)
			r.Get("/status/{jobID}", s.getInfoAboutJob)
		})

		r.Group(func(r chi.Router) {
			r.Post("/login", s.login)
			r.Post("/create", s.createUser)
		})
	})

	err := http.ListenAndServe(":8080", router)
	log.Fatal(err)
	return err
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
	s.s.Save("jobs", jobID, &storage.Job{Time: time.Now(), Status: "performed"})
	go s.sendJobToConverter(u, jobID)

	render.JSON(w, r, interfaces.JSON{"code": resp.Status, "jobID": jobID})
}

func (s Server) sendJobToConverter(u *url.URL, id string) {
	var err error
	defer func() {
		var status string
		if err == nil {
			status = "performed"
		} else {
			status = "fail"
		}
		s.s.Save("jobs", id, &storage.Job{Time: time.Now(), Status: status})
	}()

	v, err := youtube.NewFromURL(u)
	if err != nil {
		log.Printf("[%s] [WARN] error in getting info about %s\n", id, u.String())
		return
	}
	log.Printf("[%s] [INFO] v %s", id, v.Duration)

	cfg := s.cfgReader.Read()
	node := cfg.Converters.Next()
	converterURL, _ := url.Parse(node.Adress)
	converterURL.Scheme = "http"

	log.Printf("[%s] [INFO] send request to %s->%s", id, node.Name, converterURL.String())

	data, err := json.Marshal(interfaces.JSON{"job_id": id, "link": v.Formats[0].URL})
	if err != nil {
		log.Printf("[%s] [WARN] json marshal error %s", err.Error())
		return
	}

	resp, err := http.Post(converterURL.String(), "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("[WARN] error put job into queue %s", err.Error())
		return
	}
	log.Printf("[INFO] response code %d", resp.StatusCode)
}

func (s Server) getInfoAboutJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	if len(jobID) < 64 {
		render.Render(w, r, errors.InvalidRequest)
		return
	}

	log.Printf("[%s] [INFO] info about job %s", middleware.GetReqID(r.Context()), jobID)
	var job storage.Job
	s.s.Load("jobs", jobID, &job)
	render.JSON(w, r, interfaces.JSON{"status": job.Status, "jobID": jobID})
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
	_, token, err := s.token.Encode(jwtauth.Claims{"exp": now.Add(30 * time.Minute).Unix(), "reqPerHour": u.Permissions.RequestPerHour, "ttl": u.Permissions.TTL})
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
