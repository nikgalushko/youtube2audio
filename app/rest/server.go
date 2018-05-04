package rest

import (
	"log"
	"sync"

	"github.com/jetuuuu/youtube2audio/app/config"
	"github.com/jetuuuu/youtube2audio/app/rest/interfaces"
	"github.com/jetuuuu/youtube2audio/app/rest/private"
	"github.com/jetuuuu/youtube2audio/app/rest/public"
	"github.com/jetuuuu/youtube2audio/app/storage"
)

type Server struct {
	apis []interfaces.Api
}

func New(c config.ConfigReader, store *storage.Storage) Server {
	s := Server{apis: []interfaces.Api{private.New(c, store), public.New(c, store)}}
	return s
}

func (s *Server) Run() {
	var wg sync.WaitGroup
	for _, a := range s.apis {
		wg.Add(1)
		go func(a interfaces.Api) {
			log.Fatal(a.Run())
			wg.Done()
		}(a)
	}
	wg.Wait()
}
