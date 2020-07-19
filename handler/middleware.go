package handler

import (
	dterrors "shakilakhtar/go-microservices-platform/errors"
	"shakilakhtar/go-microservices-platform/security/uaa"
	logger "github.com/sirupsen/logrus"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"time"
)

// Default Access control handler chain
var DefaultAccessControlChain = []handlerFunc{
	AccessControlHandler,
}

// take in one HandlerFunc and wrap it within another HandlerFunc
type handlerFunc func(http.HandlerFunc) http.HandlerFunc

// HandlerFuncChain builds the handler functions chain recursively
func HandlerFuncChain(f http.HandlerFunc, h ...handlerFunc) http.HandlerFunc {
	// if our chain is done, use the original handlerfunc
	if len(h) == 0 {
		//logger.Debug("No handler chain found calling handler","Handler:", f)
		return f
	}
	// otherwise nest the handlerfuncs
	return h[0](HandlerFuncChain(f, h[1:cap(h)]...))
}

//Handler adapter wraps an http.Handler with additional functionality
type HandlerAdapter func(http.Handler) http.Handler

//chain handler with all specified functionality
func Adapt(h http.Handler, adapters ...HandlerAdapter) http.Handler{
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

func Notify() HandlerAdapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("before")
			defer logger.Debug("after")
			h.ServeHTTP(w, r)
		})
	}
}

func LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		logger.Debug("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

//A handler for recovering Panic
func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Debug("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func AccessControlHandler(h http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type,Authorization")

		if r.Method == "OPTIONS" {
			return
		}
		if os.Getenv("TOKEN_VALIDATION") == "true" {

			var token = r.Header.Get("Authorization")
			helper := uaa.NewUaaHelper(os.Getenv("TOKEN_VALIDATION_URL"), "", os.Getenv("Client_Credentials"))
			isValid, errResponse := helper.IsValidToken(token)
			if isValid == false {
				w.WriteHeader(errResponse.Status)
				errResp, _ := json.Marshal(errResponse)
				w.Write([]byte(errResp))
			}
			return
		}

		h.ServeHTTP(w, r)

	}
	return http.HandlerFunc(fn)
}

func BodyParserHandler(v interface{}) func(http.Handler) http.Handler {
	t := reflect.TypeOf(v)

	m := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			val := reflect.New(t).Interface()
			err := json.NewDecoder(r.Body).Decode(val)

			if err != nil {
				dterrors.WriteError(w, dterrors.ErrBadRequest)
				return
			}

			//context.Set(r, "body", val)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}

	return m
}
