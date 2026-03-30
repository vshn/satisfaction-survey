package satisfaction

import (
	"net/http"

	"github.com/vshn/satisfaction-survey/pkg/api/handler"
)

type satisfactionServer struct {
	counter ResultCounter
}

type ResultCounter interface {
	IncPositive()
	IncNegative()
}

func (s *satisfactionServer) AddPositive(r *http.Request) (any, error) {
	s.counter.IncPositive()
	return "success", nil
}

func (s *satisfactionServer) AddNegative(r *http.Request) (any, error) {
	s.counter.IncNegative()
	return "success", nil
}

func Setup(mux *http.ServeMux, counter ResultCounter) {
	s := satisfactionServer{counter}
	mux.Handle("POST /yay", handler.JSONFunc(s.AddPositive))
	mux.Handle("POST /urgh", handler.JSONFunc(s.AddNegative))
}
