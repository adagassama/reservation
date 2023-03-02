package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Shop struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Slot struct {
	ID        int       `json:"id"`
	ShopID    int       `json:"shopId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type Appointment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	SlotID    int       `json:"slotId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

var db *sql.DB

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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INT AUTO_INCREMENT PRIMARY KEY, email VARCHAR(50) UNIQUE NOT NULL, password VARCHAR(100) NOT NULL)")
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

	router.POST("/register", Register)
	router.POST("/login", Login)
	router.POST("/shops", CreateShop)
	router.POST("/slots", CreateSlot)
	router.POST("/appointments", CreateAppointment)
	router.GET("/appointments", GetAppointments)

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

func Register(c *gin.Context) {
	var user User
	user.Email = c.PostForm("email")
	user.Password = c.PostForm("password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error hashing password"})
		return
	}
	fmt.Println("Password Haché:", hashedPassword)

	_, err = db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", user.Email, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating user en GOOO"})
		return
	}

	/*userID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"userID": userID})*/
	c.Redirect(http.StatusFound, "/connexion")
}

func Login(c *gin.Context) {
	var user User
	user.Email = c.PostForm("email")
	user.Password = c.PostForm("password")

	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE email = ?", user.Email).Scan(&storedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error2": "invalid credentials"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error3": "invalid credentials"})
		return
	}
	//c.JSON(http.StatusOK, gin.H{"message": "login successful"})
	c.Redirect(http.StatusFound, "/dashboard")
}

func CreateShop(c *gin.Context) {
	var shop Shop
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
	var slot Slot
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
	var appointment Appointment
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

	/*appointments := []Appointments{}
	for rows.Next() {
		var appointment AppointmentDetails
		if err := rows.Scan(&appointment.ID, &appointment.ShopName, &appointment.ShopAddress, &appointment.StartTime, &appointment.EndTime); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error scanning appointment"})
			return
		}
		appointments = append(appointments, appointment)
	}

	c.JSON(http.StatusOK, appointments)*/
}
