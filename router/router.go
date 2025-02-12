package router

import (
	"ass3_part2/controllers"
	"ass3_part2/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter() http.Handler {
	router := mux.NewRouter()

	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.RateLimit)

	authRoutes := router.PathPrefix("/").Subrouter()
	authRoutes.Use(middleware.MiddlewareAuth)

	router.HandleFunc("/index", serveHTML("static/index.html"))

	adminRoutes := router.PathPrefix("/admin").Subrouter()
	//adminRoutes.Use(middleware.MiddlewareAuth)
	//adminRoutes.Use(middleware.MiddlewareRole("admin"))
	adminRoutes.HandleFunc("/subscription", controllers.CreateSubscription).Methods("POST")
	router.HandleFunc("/subscription/{id}", controllers.GetSubscription).Methods("GET")
	router.HandleFunc("/subscription", controllers.GetAllSubscriptions).Methods("GET")
	adminRoutes.HandleFunc("/subscription/{id}", controllers.DeleteSubscription).Methods("DELETE")
	adminRoutes.HandleFunc("/subscription/{id}", controllers.UpdateSubscription).Methods("PUT")

	router.HandleFunc("/payment", controllers.PaySubscription).Methods("POST")
	//middleware only here!

	return middleware.CORSMiddleware(router)
}
func serveHTML(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}
}
