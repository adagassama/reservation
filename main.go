package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"reservation/models"
	"time"
)

var db *sql.DB
var jwtKey = []byte("3EDC&edc")

func main() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:8889)/appointment")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	// create tables if they don't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INT AUTO_INCREMENT PRIMARY KEY, firstname VARCHAR(100) NOT NULL, lastname VARCHAR(100) NOT NULL, contact VARCHAR(20) NOT NULL, email VARCHAR(50) UNIQUE NOT NULL, password VARCHAR(100) NOT NULL, status VARCHAR(100) NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS shops (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50) NOT NULL, address VARCHAR(100) NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS slots (id INT AUTO_INCREMENT PRIMARY KEY, shop_id INT NOT NULL, start_time DATETIME NOT NULL, end_time DATETIME NOT NULL, UNIQUE (shop_id, start_time), FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS appointments (id INT AUTO_INCREMENT PRIMARY KEY, user_id INT NOT NULL, shop_id INT NOT NULL, slot_id INT NOT NULL, start_time DATETIME NOT NULL, FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE, FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE CASCADE, FOREIGN KEY (slot_id) REFERENCES slots(id) ON DELETE CASCADE)")
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.POST("/register", Register)
	router.POST("/login", Login)
	router.POST("/shops", CreateShop)
	router.POST("/slots", CreateSlot)
	router.POST("/appointments", CreateAppointment)
	router.GET("/appointments", GetAppointments)
	router.GET("/", GetShops)

	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	router.GET("/login", GetSessionErrorMsg)
	router.GET("/register", GetSessionErrorMsg1)
	router.GET("/create-shop", func(c *gin.Context) {
		c.HTML(http.StatusOK, "createShop.gohtml", gin.H{
			"title": "Register",
		})
	})
	router.GET("/dashboard", AuthMiddleware(), func(c *gin.Context) {
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
func GetSessionErrorMsg(c *gin.Context) {
	session := sessions.Default(c)
	errMsg := session.Get("errMsg")
	session.Delete("errMsg")
	session.Save()
	c.HTML(http.StatusOK, "login.gohtml", gin.H{
		"title":  "Connexion",
		"errMsg": errMsg,
	})
}

func GetSessionErrorMsg1(c *gin.Context) {
	session := sessions.Default(c)
	errMsg := session.Get("errMsg")
	session.Delete("errMsg")
	session.Save()

	c.HTML(http.StatusOK, "register.gohtml", gin.H{
		"title":  "Register",
		"errMsg": errMsg,
	})
}
func Register(c *gin.Context) {
	var user models.User
	user.Firstname = c.PostForm("firstname")
	user.Lastname = c.PostForm("lastname")
	user.Contact = c.PostForm("contact")
	user.Email = c.PostForm("email")
	user.Password = c.PostForm("password")
	user.Status = c.PostForm("status")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error hashing password"})
		return
	}

	_, err = db.Exec("INSERT INTO users (firstname, lastname, contact, email, password, status) VALUES (?, ?, ?, ?, ?, ?)", user.Firstname, user.Lastname, user.Contact, user.Email, hashedPassword, user.Status)
	if err != nil {
		session := sessions.Default(c)
		session.Set("errMsg", "Erreur de création d'utilisateur")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/register")
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func Login(c *gin.Context) {
	var user models.User
	user.Email = c.PostForm("email")
	user.Password = c.PostForm("password")

	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE email = ?", user.Email).Scan(&storedPassword)

	if err != nil {
		session := sessions.Default(c)
		session.Set("errMsg", "Votre e-mail ou mot de passe est incorrect.")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)); err != nil {
		session := sessions.Default(c)
		session.Set("errMsg", "Votre e-mail ou mot de passe est incorrect.")
		session.Save()
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	// Création du token JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		IssuedAt:  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signer le token avec la clé secrète
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	// Stocker le token dans le local storage du navigateur
	c.SetCookie("gass", tokenString, 86400, "/", "localhost", false, true)

	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("gass")
		fmt.Println(token)

		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("3EDC&edc"), nil
		})

		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func CreateShop(c *gin.Context) {
	var shop models.Shop
	shop.Name = c.PostForm("nameshop")
	shop.Address = c.PostForm("address")

	_, err := db.Exec("INSERT INTO shops (name, address) VALUES (?, ?)", shop.Name, shop.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating shop"})
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func CreateSlot(c *gin.Context) {
	var slot models.Slot
	if err := c.ShouldBindJSON(&slot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if the shop exists
	var shopID int
	err := db.QueryRow("SELECT id FROM shops WHERE id = ?", slot.ShopID).Scan(&shopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop ID"})
		return
	}

	// check if the slot overlaps with existing slots
	var overlappingSlotID int
	err = db.QueryRow("SELECT id FROM slots WHERE shop_id = ? AND ((? BETWEEN start_time AND end_time) OR (? BETWEEN start_time AND end_time))", slot.ShopID, slot.StartTime, slot.EndTime).Scan(&overlappingSlotID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "overlapping slot"})
		return
	}

	result, err := db.Exec("INSERT INTO slots (shop_id, start_time, end_time) VALUES (?, ?, ?)", slot.ShopID, slot.StartTime, slot.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating slot"})
		return
	}

	slotID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting slot ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slotID": slotID})
}

func CreateAppointment(c *gin.Context) {
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if the shop and slot exist
	var shopID int
	var startTime, endTime time.Time
	err := db.QueryRow("SELECT shops.id, slots.start_time, slots.end_time FROM shops JOIN slots ON shops.id = slots.shop_id WHERE slots.id = ?", appointment.SlotID).Scan(&shopID, &startTime, &endTime)
	/*if err != sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid slot ID"})
		return
	}*/

	// check if the slot is available
	var appointmentID int
	err = db.QueryRow("SELECT id FROM appointments WHERE slot_id = ? AND start_time = ?", appointment.SlotID, appointment.StartTime).Scan(&appointmentID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot not available"})
		return
	}

	// create the appointment
	result, err := db.Exec("INSERT INTO appointments (user_id, shop_id, slot_id, start_time) VALUES (?, ?, ?, ?)", appointment.UserID, shopID, appointment.SlotID, appointment.StartTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating appointment"})
		return
	}

	appointID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting appointment ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointmentID": appointID})
}

func GetAppointments(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
		return
	}

	rows, err := db.Query("SELECT appointments.id, shops.name, shops.address, slots.start_time, slots.end_time FROM appointments JOIN slots ON appointments.slot_id = slots.id JOIN shops ON slots.shop_id = shops.id WHERE appointments.user_id = ? ORDER BY slots.start_time ASC", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting appointments"})
		return
	}
	defer rows.Close()
	c.JSON(http.StatusOK, " GOOD ")

}
func GetShops(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM shops")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	// Stockage des données dans un slice de `Shop`
	var shops []models.Shop
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(&shop.ID, &shop.Name, &shop.Address)
		if err != nil {
			log.Fatal(err)
		}
		shops = append(shops, shop)
	}
	fmt.Println(shops)
	c.HTML(http.StatusOK, "index.gohtml", gin.H{
		"title": "Tintidale",
		"shops": shops,
	})
}
