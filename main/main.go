package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", gin.H{
			"title": "Tintidale",
		})
	})
	router.GET("/connexion", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.gohtml", gin.H{
			"title": "Connexion",
		})
	})
	router.GET("/create-account", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.gohtml", gin.H{
			"title": "Register",
		})
	})
	router.GET("/create-shop", func(c *gin.Context) {
		c.HTML(http.StatusOK, "createShop.gohtml", gin.H{
			"title": "Register",
		})
	})
	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.gohtml", gin.H{
			"title": "Dashboard",
		})
	})
	router.GET("/booking", func(c *gin.Context) {
		c.HTML(http.StatusOK, "booking.html", gin.H{
			"title": "Booking",
		})
	})
	router.Run(":2020")
}
