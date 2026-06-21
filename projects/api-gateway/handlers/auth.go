package handlers

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"net/http"
	"gopkg.in/yaml.v3"
	"os"
	"github.com/gorilla/sessions"
)

type Config struct {
	Users     []User   `yaml:"users"`
	Upstream  Upstream `yaml:"upstream"`
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Upstream struct {
	URL string `yaml:"url"`
}

var (
	store = sessions.NewCookieStore([]byte("secret-key"))
	config Config
)

func init() {
	loadConfig()
}

func loadConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func verifyPassword(hashedPassword, inputPassword string) (bool, error) {
	decodedHash, err := base64.StandardEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false, fmt.Errorf("failed to decode hashed password: %v", err)
	}

	salt := []byte("saltysalt")
	hash := argon2.IDKey([]byte(inputPassword), salt, 3, 64*1024, 2, 32)

	return subtle.ConstantTimeCompare(decodedHash, hash) == 1, nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		for _, user := range config.Users {
			if user.Username == username {
				match, err := verifyPassword(user.Password, password)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				if match {
					session, _ := store.Get(r, "session")
					session.Values["authenticated"] = true
					session.Save(r, w)
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	http.ServeFile(w, r, "templates/login.html")
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}