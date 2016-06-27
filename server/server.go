package server

import (
	"bytes"
	"github.com/vil-coyote-acme/go-concurrency/commons"
	"log"
	"net/http"
	"encoding/json"
	"io"
)

const bartenderPath string = "/orders"

func NewServer(playerId string, bartenderUrl string) (s *Server) {
	s = new(Server)
	s.bartenderUrl = bartenderUrl
	s.playerId = playerId
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/orders", s.handleOrder)
	return s
}

type Server struct {
	playerId     string
	bartenderUrl string
	mux          *http.ServeMux
}

func (s *Server) Start(url string) {
	err := http.ListenAndServe(url, s.mux)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (s *Server) handleOrder(w http.ResponseWriter, r *http.Request) {
	// first step : unmarshal the incoming order
	var order commons.Order
	unMarshallErr := unmarshalOrderFromHttp(r, &order)
	if unMarshallErr != nil {
		log.Println(unMarshallErr.Error())
		return
	}
	log.Printf("receive one order : %s", order)

	order.PlayerId = s.playerId
	buf, marshalErr := json.Marshal(order)
	if marshalErr != nil {
		log.Println(marshalErr.Error())
		return
	}
	// second step, send the order to the bartender
	res, err := http.Post(s.bartenderUrl + bartenderPath, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Printf("error when calling bartender api : %s", err)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("get a non 200 response when calling bartender api : %s", res.Status)
		return
	}

	// third step, if all is right, get your money back !
	paymentRes, paymentErr := http.Get(order.CallBackUrl)
	if paymentErr != nil {
		log.Printf("get an error when calling payment api : %s", paymentErr.Error())
	}
	if paymentRes != nil && paymentRes.StatusCode != 200 {
		log.Printf("get a non 200 response when calling payment api : %s", paymentRes.Status)
	}
	w.WriteHeader(200)
}

func unmarshalOrderFromHttp(r *http.Request, order *commons.Order) (err error) {
	buf := make([]byte, r.ContentLength)
	io.ReadFull(r.Body, buf)
	err = json.Unmarshal(buf, &order)
	return
}
