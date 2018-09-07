package main

import (
	"crypto/rsa"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/mrjones/oauth"
	"github.com/rs/zerolog"
	"gopkg.in/redis.v5"
)

type Settings struct {
	SecretKey   string `envconfig:"SECRET_KEY" required:"true"`
	Hostname    string `envconfig:"HOSTNAME" required:"true"`
	Port        string `envconfig:"PORT" required:"true"`
	PostgresURL string `envconfig:"DATABASE_URL" required:"true"`
	RedisURL    string `envconfig:"REDIS_URL"`
	PrivateKey  string `envconfig:"PRIVATE_KEY" required:"true"`
	PublicKey   string `envconfig:"PUBLIC_KEY" required:"true"`
}

var err error
var s Settings
var r *mux.Router
var c *oauth.Consumer
var pg *sqlx.DB
var rds *redis.Client
var store sessions.Store
var privateKey *rsa.PrivateKey
var log = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr})

func main() {
	err = envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig.")
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log = log.With().Timestamp().Logger()

	// cookie store
	store = sessions.NewCookieStore([]byte(s.SecretKey))

	// keys
	s.PublicKey = strings.Replace(s.PublicKey, "\\n", "\n", -1)
	s.PrivateKey = strings.Replace(s.PrivateKey, "\\n", "\n", -1)
	privateKey, err = decodePrivateKeyPEM([]byte(s.PrivateKey))
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't decode private key")
	}

	// postgres connection
	pg, err = sqlx.Connect("postgres", s.PostgresURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to postgres")
	}

	// redis connection
	if s.RedisURL != "" {
		rurl, _ := url.Parse(s.RedisURL)
		pw, _ := rurl.User.Password()
		rds = redis.NewClient(&redis.Options{
			Addr:     rurl.Host,
			Password: pw,
		})

		if err := rds.Ping().Err(); err != nil {
			log.Fatal().Err(err).Str("url", s.RedisURL).
				Msg("failed to connect to redis")
		}
	}

	// define routes
	r = mux.NewRouter()
	r.Path("/favicon.ico").Methods("GET").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./public/icon.png")
			return
		})

	r.Path("/public-key").Methods("GET").HandlerFunc(servePublicKey)
	r.Path("/transfer").Methods("POST").HandlerFunc(receiveTransfer)
	r.Path("/transfer/ack").Methods("POST").HandlerFunc(receiveTransferAck)

	// start the server
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + s.Port,
		WriteTimeout: 25 * time.Second,
		ReadTimeout:  25 * time.Second,
	}
	log.Info().Str("port", s.Port).Msg("listening.")
	srv.ListenAndServe()
}
