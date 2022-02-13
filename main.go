package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func getCakeHandler(w http.ResponseWriter, r *http.Request, u User) {
	startTime := time.Now()

	w.Write([]byte(u.FavoriteCake))
	duration := time.Since(startTime)
	responseTimeHistogram.WithLabelValues("/cake").Observe(duration.Seconds())

	// getCakeTime.Observe(float64())
	numberOfCakesGiven.Inc()
}

func main() {
	os.Setenv("CAKE_ADMIN_EMAIL", "admin@mail.com")
	os.Setenv("CAKE_ADMIN_PASSWORD", "admin1111")

	go send()
	go metrics()

	flag.Parse()
	hub := newHub()
	go hub.run()
	r := mux.NewRouter()
	users := NewInMemoryUserStorage()
	userService := UserService{repository: users}

	jwtService, err := NewJWTService("pubkey.rsa", "privkey.rsa")
	if err != nil {
		panic(err)
	}

	r.HandleFunc("/cake", logRequest(jwtService.jwtAuth(users, getCakeHandler))).Methods(http.MethodGet)
	r.HandleFunc("/user/register", logRequest(userService.Register)).Methods(http.MethodPost)
	r.HandleFunc("/user/jwt", logRequest(wrapJwt(jwtService, userService.JWT))).Methods(http.MethodPost)

	r.HandleFunc("/user/me", logRequest(userService.Profile)).Methods(http.MethodGet)
	r.HandleFunc("/user/favorite_cake", logRequest(userService.ChangeCake)).Methods(http.MethodPost)
	r.HandleFunc("/user/email", logRequest(userService.ChangeEmail)).Methods(http.MethodPost)
	r.HandleFunc("/user/password", logRequest(userService.ChangePassword)).Methods(http.MethodPost)

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r, jwtService, *users)
	})

	srv := http.Server{
		Addr:    ":8000",
		Handler: r,
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {

		<-interrupt
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Println("Server started, hit Ctrl+C to stop")
	err = srv.ListenAndServe()
	if err != nil {
		log.Println("Server exited with error:", err)
	}
	log.Println("Good bye :)")
}

func wrapJwt(jwt *JWTService, f func(http.ResponseWriter, *http.Request, *JWTService)) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f(rw, r, jwt)
	}
}

