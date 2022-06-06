package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/mwettste/greenlight/internal/data"
	"github.com/mwettste/greenlight/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				writer.Header().Set("Connection", "close")
				app.serverErrorResponse(writer, request, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(writer, request)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					fmt.Printf("deleting ip '%s' from client cache\n", ip)
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	totalRateLimiterHits := expvar.NewInt("total_rate_limiter_hits")

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if app.config.limiter.enabled {
			ip := realip.FromRequest(request)

			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				totalRateLimiterHits.Add(1)
				app.rateLimitExceededResponse(writer, request)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(writer, request)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user := app.contextGetUser(request)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(writer, request)
			return
		}

		if !user.Activated {
			app.inactiveAccountResponse(writer, request)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user := app.contextGetUser(request)

		if !user.Activated {
			app.inactiveAccountResponse(writer, request)
			return
		}

		next.ServeHTTP(writer, request)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(writer http.ResponseWriter, request *http.Request) {
		user := app.contextGetUser(request)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(writer, request, err)
			return
		}

		if !permissions.Includes(code) {
			app.notPermittedResponse(writer, request)
			return
		}

		next.ServeHTTP(writer, request)
	}

	return app.requireActivatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Vary", "Origin")
		writer.Header().Add("Vary", "Access-Control-Request-Method")

		origin := request.Header.Get("Origin")
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					writer.Header().Set("Access-Control-Allow-Origin", origin)

					if request.Method == http.MethodOptions && request.Header.Get("Access-Control-Request-Method") != "" {
						writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						writer.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(writer, request)
	})
}

func (application *application) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_us")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_by_status")

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		totalRequestsReceived.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, writer, request)
		totalResponsesSent.Add(1)
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
