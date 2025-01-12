package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	// "strings"
	"time"
	"weight-challenge/models"
	"math"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	// "net"
	// "runtime/debug"
)

var db *sql.DB

func setupCORS(r *gin.Engine) {
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("API_URL"))
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })
}

func main() {
	var err error

	// Зареждане на .env файл
	if err = godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Създаваме директория за логове ако не съществува
	if err = os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Error creating logs directory:", err)
	}

	// Отваряме файл за логове
	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer logFile.Close()

	// Конфигурираме логването според средата
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "development" {
		// В development режим логваме във файл и в конзолата
		log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	} else {
		// В production режим логваме само във файл
		log.SetOutput(logFile)
	}
	
	// Важно: задаваме формат на логовете да включва файл и ред
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Свързване с базата данни
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_CHARSET"))

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
    // Задаваме UTF-8 локализация
    os.Setenv("LANG", "bg_BG.UTF-8")
    os.Setenv("LC_ALL", "bg_BG.UTF-8")
    
    // Конфигурираме логването да използва UTF-8
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    
	// Проверка на връзката
	if err = db.Ping(); err != nil {
		log.Fatal("Could not connect to database:", err)
	}

	log.Printf("Server starting in %s mode", appEnv)
	log.Println("Successfully connected to database")
	defer db.Close()

	r := gin.New()

	setupCORS(r)

	// Добавяме middleware за UTF-8 хедъри
	// r.Use(func(c *gin.Context) {
	// 	c.Header("Content-Type", "application/json; charset=utf-8")
	// 	c.Next()
	// })

	// Конфигурираме Gin logger според средата
	if appEnv == "development" {
		r.Use(gin.Logger())
	}
	r.Use(gin.Recovery()) // Add recovery middleware separately

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// Автентикация
	r.POST("/register", register)
	r.POST("/login", login)
	r.POST("/reset-password", resetPassword)
	r.GET("/auth-status", authStatus)
	r.POST("/logout", logout)
	// Защитени endpoints
	authorized := r.Group("/")
	authorized.Use(authMiddleware())
	{
		// Тегло
		authorized.POST("/api/weight", addWeight)
		authorized.GET("/api/weight/stats", getWeightStats)
		authorized.DELETE("/api/weight/:id", deleteWeight)

		// Настройки
		authorized.GET("/api/settings", getUserSettings)
		authorized.PUT("/api/settings", updateUserSettings)
		authorized.PUT("/api/settings/visibility", updateVisibility)
		authorized.PUT("/api/user/password", changePassword)

		// Потребители
		authorized.GET("/api/users", getVisibleUsers)

		// Приятелства
		authorized.GET("/api/friends", getFriends)
		authorized.POST("/api/friends/request/:userId", sendFriendRequest)
		authorized.POST("/api/friends/accept/:friendshipId", acceptFriendRequest)
		authorized.POST("/api/friends/reject/:friendshipId", rejectFriendRequest)

		// Съревнования
		authorized.GET("/api/challenges", getChallenges)
		authorized.POST("/api/challenges", createChallenge)
		authorized.PUT("/api/challenges/:challengeId/accept", acceptChallenge)
		authorized.PUT("/api/challenges/:challengeId/reject", rejectChallenge)
		authorized.GET("/api/challenges/:challengeId/results", getChallengeResults)

		// Калории
		authorized.GET("/api/calories/stats", getCalorieStats)
		authorized.GET("/api/calories/settings", getCalorieSettings)
		authorized.PUT("/api/calories/settings", updateCalorieSettings)
		authorized.GET("/api/calories/calculations", getCalorieCalculations)
		authorized.GET("/api/calories/log/", getDailyCalorieLog)
		authorized.POST("/api/calories/food", addFoodEntry)
		authorized.POST("/api/calories/activity", addActivityEntry)
		authorized.DELETE("/api/calories/food/:id", deleteFoodEntry)
		authorized.DELETE("/api/calories/activity/:id", deleteActivityEntry)
		authorized.GET("/api/calories/meal-types", getMealTypes)
		authorized.GET("/api/calories/needs", calculateCalorieNeeds)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://%s:%s", os.Getenv("API_URL"), port)
	r.Run(":" + port)
}

func logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func register(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Валидация
	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Debug logging
	log.Printf("Registration attempt - Username: %s, Password length: %d",
		user.Username, len(user.Password))

	// Проверка дали потребителят вече съществува
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)",
		user.Username).Scan(&exists)
	if err != nil {
		log.Printf("Database error checking user existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Хеширане на паролата
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	// Запис в базата
	result, err := db.Exec(`
        INSERT INTO users (username, password, height) 
        VALUES (?, ?, ?)`,
		user.Username, string(hashedPassword), user.Height)
	if err != nil {
		log.Printf("Database error inserting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	user.Password = "" // Не връщаме паролата

	log.Printf("Successfully registered user: ID=%d, Username=%s", user.ID, user.Username)

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
		"user":    user,
	})
}

func login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&credentials); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Валидация
	if credentials.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if credentials.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Debug logging
	log.Printf("Login attempt - Username: %s, Password length: %d",
		credentials.Username, len(credentials.Password))

	var user models.User
	var hashedPassword string
	err := db.QueryRow(`
        SELECT id, username, password, height, age, gender
        FROM users 
        WHERE username = ?`,
		credentials.Username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Height, &user.Age, &user.Gender)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", credentials.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Проверка на паролата
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
	if err != nil {
		log.Printf("Invalid password for user %s: %v", credentials.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Създаваме токен с потребителското ID
	// token := fmt.Sprintf("user-%d", user.ID)
	tokenDetails, err := createToken(user.ID)
	if err != nil {
		log.Printf("Error creating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}
	token := tokenDetails.AccessToken
	log.Printf("Token details: %+v", tokenDetails)

	log.Printf("Successful login for user: %s", credentials.Username)

	// c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("access_token", token, int(tokenDetails.AtExpires), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

func authStatus(c *gin.Context) {
	tokenString, err := c.Cookie("access_token")

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "isAuthenticated": false,
            "error": "No token found",
        })
        return
    }

    // Парсване на token без claims
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("ACCESS_SECRET")), nil
    })

	fmt.Println("token", token)

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "isAuthenticated": false,
            "error": "Invalid token authStatus",
        })
        return
    }

    if !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{
            "isAuthenticated": false,
            "error": "Token is not valid",
        })
        return
    }
	user := models.User{
		ID:       int(token.Claims.(jwt.MapClaims)["user_id"].(float64)),
		Username: token.Claims.(jwt.MapClaims)["username"].(string),
		Height:   token.Claims.(jwt.MapClaims)["height"].(float64),
	}

    c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"user": user,
        "isAuthenticated": true,
    })
}

type TokenDetails struct {
    AccessToken  string
    RefreshToken string
    AccessUuid   string
    RefreshUuid  string
    AtExpires    int64
    RtExpires    int64
}

func createToken(userID int) (*TokenDetails, error) {
    td := &TokenDetails{}
    td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
    td.AccessUuid = uuid.New().String()

    td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
    td.RefreshUuid = uuid.New().String()

	row := db.QueryRow("SELECT id, username, height FROM users WHERE id = ?", userID)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Height)
	if err != nil {
		return nil, err
	}
	log.Printf("User: %+v", user)
    // Access Token
    atClaims := jwt.MapClaims{}
    atClaims["authorized"] = true
    atClaims["access_uuid"] = td.AccessUuid
    atClaims["user_id"] = user.ID
	atClaims["username"] = user.Username
	atClaims["height"] = user.Height
    atClaims["exp"] = td.AtExpires

    at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
    td.AccessToken, _ = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

    // Refresh Token
    rtClaims := jwt.MapClaims{}
    rtClaims["refresh_uuid"] = td.RefreshUuid
    rtClaims["user_id"] = userID
	rtClaims["username"] = user.Username
	rtClaims["height"] = user.Height
    rtClaims["exp"] = td.RtExpires

    rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
    td.RefreshToken, _ = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))

    return td, nil
}

func CheckPasswordHash(password, storedHash string) bool {
	// Debug информация
	log.Printf("Checking password: %s", password)
	log.Printf("Against stored hash: %s", storedHash)
	log.Printf("Password length: %d", len(password))
	log.Printf("Stored hash length: %d", len(storedHash))

	// Проверка на хеша
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		log.Printf("Password check failed: %v", err)
		return false
	}

	return true
}

func addWeight(c *gin.Context) {
	var input models.WeightRecordInput
	if err := c.BindJSON(&input); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)

	// Парсване на датата
	createdAt, err := time.Parse(time.RFC3339, input.CreatedAt)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	record := models.WeightRecord{
		UserID:    userID,
		Weight:    input.Weight,
		CreatedAt: createdAt,
	}

	result, err := db.Exec("INSERT INTO weight_records (user_id, weight, created_at) VALUES (?, ?, ?)",
		record.UserID, record.Weight, record.CreatedAt)
	if err != nil {
		log.Printf("Error saving weight record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save weight record"})
		return
	}

	id, _ := result.LastInsertId()
	record.ID = int(id)

	c.JSON(http.StatusOK, record)
}

func getWeightStats(c *gin.Context) {
	userID := getUserID(c)

	var stats models.WeightStats

	// Вземаме височината на потребителя
	err := db.QueryRow("SELECT height FROM users WHERE id = ?", userID).Scan(&stats.Height)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch user data"})
		return
	}

	// Вземаме всички записи, сортирани по дата
	rows, err := db.Query(`
		SELECT id, weight, created_at 
		FROM weight_records 
		WHERE user_id = ? 
		ORDER BY created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch weight records"})
		return
	}
	defer rows.Close()

	var records []models.WeightRecord
	for rows.Next() {
		var record models.WeightRecord
		if err := rows.Scan(&record.ID, &record.Weight, &record.CreatedAt); err != nil {
			continue
		}
		records = append(records, record)
	}

	if len(records) > 0 {
		stats.CurrentWeight = records[0].Weight
		stats.InitialWeight = records[len(records)-1].Weight
		stats.TotalProgress = models.CalculateProgress(stats.InitialWeight, stats.CurrentWeight)
		stats.BMI = models.CalculateBMI(stats.CurrentWeight, stats.Height)

		if len(records) > 1 {
			stats.PreviousWeight = records[1].Weight
			stats.DailyProgress = models.CalculateProgress(stats.PreviousWeight, stats.CurrentWeight)
		}
	}

	stats.History = records

	c.JSON(http.StatusOK, stats)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Request path: %s", c.Request.URL.Path)
		// Check if request is from localhost
		// remoteAddr := c.Request.RemoteAddr
		// host, _, _ := net.SplitHostPort(remoteAddr)
		// requestHost := c.Request.Host
		
		// Check both RemoteAddr and Host header for localhost
		/*isLocalhost := host == "127.0.0.1" || host == "::1" || host == "localhost" ||
			strings.HasPrefix(requestHost, "localhost:") || strings.HasPrefix(requestHost, "127.0.0.1:") || strings.HasPrefix(requestHost, "[::1]:")
		
		if isLocalhost {
			// For localhost requests, set a default user ID of 1
			c.Set("userID", 1)
			c.Next()
			return
		}*/

		// Извличаме ID-то от cookie
		cookie, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("ACCESS_SECRET")), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Извличаме ID-то от токена
		userID := int(token.Claims.(jwt.MapClaims)["user_id"].(float64))

		// Запазваме ID-то в контекста
		c.Set("userID", int(userID))
		c.Next()
	}
}

func getUserID(c *gin.Context) int {
	// Взимаме ID-то от контекста
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(int)
}

func getUserSettings(c *gin.Context) {
	log.Printf("Търсене на настройки за потребител") // Debug log
	userID := getUserID(c)

	var user models.User
	err := db.QueryRow(`
        SELECT u.id, u.username, u.first_name, u.last_name, u.age, u.height, u.gender, u.email, u.target_weight, us.is_visible
        FROM users u
        LEFT JOIN user_settings us ON u.id = us.user_id
        WHERE u.id = ?`, userID).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName,
		&user.Age, &user.Height, &user.Gender, &user.Email, &user.Target, &user.IsVisible)

	if err != nil {
		log.Printf("Error fetching user settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch user settings"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func updateUserSettings(c *gin.Context) {
	userID := getUserID(c)
	var settings models.User

	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	_, err := db.Exec(`
        UPDATE users 
        SET first_name = ?, last_name = ?, age = ?, height = ?, 
            gender = ?, email = ?, target_weight = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?`,
		settings.FirstName, settings.LastName, settings.Age, settings.Height,
		settings.Gender, settings.Email, settings.Target, userID)

	if err != nil {
		log.Printf("Error updating user settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

func changePassword(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Проверка на текущата парола
	var storedHash string
	err := db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&storedHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify current password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Хеширане и запазване на новата парола
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process new password"})
		return
	}

	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func getVisibleUsers(c *gin.Context) {
	userID := getUserID(c)

	rows, err := db.Query(`
        SELECT u.id, u.username, u.height, 
               COALESCE(
                   (SELECT ((w1.weight - w2.weight) / w1.weight * 100)
                    FROM weight_records w1
                    JOIN weight_records w2 ON w2.user_id = u.id
                    WHERE w1.user_id = u.id
                    AND w1.created_at = (SELECT MIN(created_at) FROM weight_records WHERE user_id = u.id)
                    AND w2.created_at = (SELECT MAX(created_at) FROM weight_records WHERE user_id = u.id)
                    LIMIT 1
                   ), 0
               ) as progress
        FROM users u
        JOIN user_settings us ON u.id = us.user_id
        WHERE us.is_visible = true 
        AND u.id != ?
        AND NOT EXISTS (
            SELECT 1 FROM friendships f
            WHERE ((f.requester_id = ? AND f.addressee_id = u.id) 
               OR (f.requester_id = u.id AND f.addressee_id = ?))
            AND f.status IN ('accepted', 'pending')
        )`,
		userID, userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}
	defer rows.Close()

	var users []models.UserProfile
	for rows.Next() {
		var user models.UserProfile
		if err := rows.Scan(&user.ID, &user.Username, &user.Height, &user.Progress); err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func updateVisibility(c *gin.Context) {
	userID := getUserID(c)
	var settings struct {
		IsVisible bool `json:"isVisible"`
	}

	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	_, err := db.Exec(`
        INSERT INTO user_settings (user_id, is_visible) 
        VALUES (?, ?)
        ON DUPLICATE KEY UPDATE is_visible = ?`,
		userID, settings.IsVisible, settings.IsVisible)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update visibility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visibility updated"})
}

func getFriends(c *gin.Context) {
	userID := getUserID(c)

	rows, err := db.Query(`
        SELECT u.id, u.username, u.height, f.status, f.id as friendship_id,
               COALESCE(
                   (SELECT ((w1.weight - w2.weight) / w1.weight * 100)
                    FROM weight_records w1
                    JOIN weight_records w2 ON w2.user_id = u.id
                    WHERE w1.user_id = u.id
                    AND w1.created_at = (SELECT MIN(created_at) FROM weight_records WHERE user_id = u.id)
                    AND w2.created_at = (SELECT MAX(created_at) FROM weight_records WHERE user_id = u.id)
                    LIMIT 1
                   ), 0
               ) as progress
        FROM users u
        JOIN friendships f ON (f.requester_id = u.id OR f.addressee_id = u.id)
        WHERE (f.requester_id = ? OR f.addressee_id = ?)
        AND u.id != ?`,
		userID, userID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch friends"})
		return
	}
	defer rows.Close()

	var friends []struct {
		models.UserProfile
		Status       string `json:"status"`
		FriendshipID int    `json:"friendshipId"`
	}

	for rows.Next() {
		var friend struct {
			models.UserProfile
			Status       string `json:"status"`
			FriendshipID int    `json:"friendshipId"`
		}
		if err := rows.Scan(&friend.ID, &friend.Username, &friend.Height, &friend.Status, &friend.FriendshipID, &friend.Progress); err != nil {
			continue
		}
		friends = append(friends, friend)
	}

	c.JSON(http.StatusOK, friends)
}

func sendFriendRequest(c *gin.Context) {
	userID := getUserID(c)
	friendIDStr := c.Param("userId")

	// Проверяваме дали ID-то не е празно
	if friendIDStr == "" {
		log.Printf("Липсващо ID на потребител в заявката")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не е предоставено ID на потребител"})
		return
	}

	// Конвертираме friendID в число
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		log.Printf("Невалидно ID на потребител: %v (стойност: %s)", err, friendIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Невалидно ID на потребител"})
		return
	}

	// Проверяваме дали потребителят не се опитва да добави себе си
	if friendID == userID {
		log.Printf("Потребител %d се опитва да добави себе си за приятел", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не можете да добавите себе си за приятел"})
		return
	}

	// Debug logging
	log.Printf("Получена заявка за приятелство от потребител %d към потребител %d", userID, friendID)

	// Проверяваме дали потребителят съществува
	var exists bool
	err = db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM users 
            WHERE id = ?
        )`, friendID).Scan(&exists)

	if err != nil {
		log.Printf("Грешка при проверка на потребителя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при проверка на потребителя"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Потребителят не е намерен"})
		return
	}

	// Проверяваме дали вече има активна или изчакваща заявка между тези потребители
	err = db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM friendships 
            WHERE ((requester_id = ? AND addressee_id = ?) 
               OR (requester_id = ? AND addressee_id = ?))
            AND status IN ('pending', 'accepted')
        )`, userID, friendID, friendID, userID).Scan(&exists)

	if err != nil {
		log.Printf("Грешка при проверка за съществуваща заявка: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при проверка за съществуваща заявка"})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Вече съществува активна заявка за приятелство между тези потребители"})
		return
	}

	// Първо проверяваме дали има отхвърлена заявка и я актуализираме
	result, err := db.Exec(`
        UPDATE friendships 
        SET status = 'pending', requester_id = ?, addressee_id = ?, updated_at = CURRENT_TIMESTAMP
        WHERE (requester_id = ? AND addressee_id = ?) 
           OR (requester_id = ? AND addressee_id = ?)
           AND status = 'rejected'`,
        userID, friendID, userID, friendID, friendID, userID)

	if err != nil {
		log.Printf("Грешка при обновяване на съществуваща заявка: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при обновяване на заявката"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	
	// Ако няма отхвърлена заявка за обновяване, създаваме нова
	if rowsAffected == 0 {
		_, err = db.Exec(`
            INSERT INTO friendships (requester_id, addressee_id) 
            VALUES (?, ?)`,
            userID, friendID)

		if err != nil {
			log.Printf("Грешка при създаване на нова заявка за приятелство: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при създаване на заявката"})
			return
		}
	}

	log.Printf("Успешно създадена/обновена заявка за приятелство от %d към %d", userID, friendID)
	c.JSON(http.StatusOK, gin.H{"message": "Заявката за приятелство е изпратена успешно"})
}

func acceptFriendRequest(c *gin.Context) {
	userID := getUserID(c)
	friendshipID := c.Param("friendshipId")

	// Проверяваме дали приятелството съществува
	var exists bool
	err := db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM friendships 
            WHERE id = ?
        )`, friendshipID).Scan(&exists)

	if err != nil {
		log.Printf("Грешка при проверка на приятелството: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при проверка на приятелството"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Приятелството не съществува"})
		return
	}

	// Проверяваме дали потребителят има право да приеме това приятелство
	var isAddressee bool
	err = db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM friendships 
            WHERE id = ? AND addressee_id = ? AND status = 'pending'
        )`, friendshipID, userID).Scan(&isAddressee)

	if err != nil {
		log.Printf("Грешка при проверка на правата: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при проверка на правата"})
		return
	}

	if !isAddressee {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нямате право да приемете това приятелство или то вече е обработено"})
		return
	}

	result, err := db.Exec(`
        UPDATE friendships 
        SET status = 'accepted' 
        WHERE id = ? AND addressee_id = ? AND status = 'pending'`,
        friendshipID, userID)

	if err != nil {
		log.Printf("Грешка при приемане на приятелството: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Възникна грешка при приемане на приятелството"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Приятелството не може да бъде прието"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Приятелството е прието успешно"})
}

func rejectFriendRequest(c *gin.Context) {
	userID := getUserID(c)
	friendshipID := c.Param("friendshipId")

	result, err := db.Exec(`
        UPDATE friendships 
        SET status = 'rejected' 
        WHERE id = ? AND addressee_id = ?`,
		friendshipID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not reject friend request"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request rejected"})
}

func createChallenge(c *gin.Context) {
	userID := getUserID(c)
	var challenge models.Challenge

	if err := c.BindJSON(&challenge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверяваме дали са приятели
	var areFriends bool
	err := db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM friendships 
            WHERE status = 'accepted' 
            AND ((requester_id = ? AND addressee_id = ?) 
                OR (requester_id = ? AND addressee_id = ?))
        )`,
		userID, challenge.OpponentID,
		challenge.OpponentID, userID).Scan(&areFriends)

	if err != nil || !areFriends {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only challenge friends"})
		return
	}

	result, err := db.Exec(`
        INSERT INTO challenges (creator_id, opponent_id, start_date, end_date) 
        VALUES (?, ?, ?, ?)`,
		userID, challenge.OpponentID, challenge.StartDate, challenge.EndDate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create challenge"})
		return
	}

	challengeID, _ := result.LastInsertId()

	// Записваме началното тегло на създателя
	var initialWeight float64
	err = db.QueryRow(`
        SELECT weight FROM weight_records 
        WHERE user_id = ? 
        ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&initialWeight)

	if err == nil {
		_, err = db.Exec(`
            INSERT INTO challenge_results (challenge_id, user_id, initial_weight) 
            VALUES (?, ?, ?)`,
			challengeID, userID, initialWeight)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Challenge created",
		"challengeId": challengeID,
	})
}

func acceptChallenge(c *gin.Context) {
	userID := getUserID(c)
	challengeID := c.Param("challengeId")

	// Проверяваме дали предизвикателството съществува и е за този потребител
	var challenge models.Challenge
	err := db.QueryRow(`
        SELECT id, creator_id, opponent_id, status 
        FROM challenges 
        WHERE id = ?`,
		challengeID).Scan(&challenge.ID, &challenge.CreatorID, &challenge.OpponentID, &challenge.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Проверяваме дали потребителят е получателят на предизвикателството
	if challenge.OpponentID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only accept challenges sent to you"})
		return
	}

	// Проверяваме дали предизвикателството е в изчакващо състояние
	if challenge.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Challenge cannot be accepted"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
		return
	}

	// Обновяваме статуса на предизвикателството
	_, err = tx.Exec(`
        UPDATE challenges 
        SET status = 'active' 
        WHERE id = ?`,
		challengeID)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update challenge"})
		return
	}

	// Опитваме се да вземем последното тегло, ако има такова
	var initialWeight float64
	err = db.QueryRow(`
        SELECT weight FROM weight_records 
        WHERE user_id = ? 
        ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&initialWeight)

	// Записваме началното тегло само ако има такова
	if err == nil {
		_, err = tx.Exec(`
            INSERT INTO challenge_results (challenge_id, user_id, initial_weight) 
            VALUES (?, ?, ?)`,
			challengeID, userID, initialWeight)

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save initial weight"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Challenge accepted"})
}

func rejectChallenge(c *gin.Context) {
	userID := getUserID(c)
	challengeID := c.Param("challengeId")

	result, err := db.Exec(`
        UPDATE challenges 
        SET status = 'rejected' 
        WHERE id = ? AND opponent_id = ? AND status = 'pending'`,
		challengeID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not reject challenge"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid challenge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Challenge rejected"})
}

func getChallenges(c *gin.Context) {
	userID := getUserID(c)

	// Debug log
	log.Printf("Fetching challenges for user ID: %d", userID)

	rows, err := db.Query(`
        SELECT c.id, c.creator_id, c.opponent_id, c.start_date, c.end_date, c.status, c.created_at,
               creator.username as creator_name, opponent.username as opponent_name
        FROM challenges c
        JOIN users creator ON c.creator_id = creator.id
        JOIN users opponent ON c.opponent_id = opponent.id
        WHERE c.creator_id = ? OR c.opponent_id = ?
        ORDER BY c.created_at DESC
    `, userID, userID)
	if err != nil {
		log.Printf("Error querying challenges: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при зареждане на предизвикателствата"})
		return
	}
	defer rows.Close()

	var challenges []models.Challenge
	for rows.Next() {
		var challenge models.Challenge
		err := rows.Scan(
			&challenge.ID,
			&challenge.CreatorID,
			&challenge.OpponentID,
			&challenge.StartDate,
			&challenge.EndDate,
			&challenge.Status,
			&challenge.CreatedAt,
			&challenge.CreatorName,
			&challenge.OpponentName,
		)
		if err != nil {
			log.Printf("Error scanning challenge: %v", err)
			continue
		}

		// Debug log
		log.Printf("Loaded challenge: %+v", challenge)

		challenges = append(challenges, challenge)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating challenges: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на предизвикателствата"})
		return
	}

	// Debug log
	log.Printf("Returning %d challenges", len(challenges))

	c.JSON(http.StatusOK, challenges)
}

func getChallengeResults(c *gin.Context) {
	userID := getUserID(c)
	challengeID := c.Param("challengeId")

	// Проверяваме дали потребителят участва в това предизвикателство
	var challenge models.Challenge
	err := db.QueryRow(`
        SELECT c.id, c.creator_id, c.opponent_id, c.start_date, c.end_date, c.status, c.created_at,
               u1.username as creator_name, u2.username as opponent_name
        FROM challenges c
        JOIN users u1 ON c.creator_id = u1.id
        JOIN users u2 ON c.opponent_id = u2.id
        WHERE c.id = ? AND (c.creator_id = ? OR c.opponent_id = ?)`,
		challengeID, userID, userID).Scan(
		&challenge.ID, &challenge.CreatorID, &challenge.OpponentID,
		&challenge.StartDate, &challenge.EndDate, &challenge.Status, &challenge.CreatedAt,
		&challenge.CreatorName, &challenge.OpponentName)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Challenge not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Вземаме резултатите за всеки участник
	rows, err := db.Query(`
        WITH user_weights AS (
            SELECT 
                u.id as user_id,
                u.username,
                COALESCE(
                    (SELECT weight 
                     FROM weight_records 
                     WHERE user_id = u.id 
                     AND created_at <= ? 
                     ORDER BY created_at ASC 
                     LIMIT 1), 0
                ) as initial_weight,
                COALESCE(
                    (SELECT weight 
                     FROM weight_records 
                     WHERE user_id = u.id 
                     AND created_at <= ? 
                     ORDER BY created_at DESC 
                     LIMIT 1), 0
                ) as final_weight
            FROM users u
            WHERE u.id IN (?, ?)
        )
        SELECT 
            user_id,
            username,
            initial_weight,
            final_weight,
            CASE 
                WHEN initial_weight > 0 AND final_weight > 0 
                THEN ((initial_weight - final_weight) / initial_weight * 100)
                ELSE 0 
            END as progress
        FROM user_weights`,
		challenge.EndDate, challenge.EndDate, challenge.CreatorID, challenge.OpponentID)

	if err != nil {
		log.Printf("Error fetching results: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch results"})
		return
	}
	defer rows.Close()

	challenge.Results = make([]models.ChallengeResult, 0)
	for rows.Next() {
		var result models.ChallengeResult
		err := rows.Scan(
			&result.UserID,
			&result.Username,
			&result.InitialWeight,
			&result.FinalWeight,
			&result.Progress,
		)
		if err != nil {
			log.Printf("Error scanning result: %v", err)
			continue
		}
		result.ChallengeID = challenge.ID
		challenge.Results = append(challenge.Results, result)
	}

	log.Printf("Challenge results: %+v", challenge) // Debug log
	c.JSON(http.StatusOK, challenge)
}

func resetPassword(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверяваме дали потребителят съществува
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Генерираме нова случайна парола
	newPassword := fmt.Sprintf("Reset%d", time.Now().Unix())

	// Хеширане на новата парола
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate new password"})
		return
	}

	// Обновяваме паролата в базата
	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Password has been reset",
		"newPassword": newPassword,
	})
}

func deleteWeight(c *gin.Context) {
	userID := getUserID(c)
	weightID := c.Param("id")

	// Първо проверяваме дали това тегло принадлежи на текущия потребител
	var ownerID int
	err := db.QueryRow("SELECT user_id FROM weight_records WHERE id = ?", weightID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Записът не е намерен"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нямате право да изтриете този запис"})
		return
	}

	// Изтриваме записа
	result, err := db.Exec("DELETE FROM weight_records WHERE id = ? AND user_id = ?", weightID, userID)
	if err != nil {
		log.Printf("Error deleting weight record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete weight record"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Записът не е намерен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Записът е изтрит успешно"})
}


func getCalorieSettingsInternal(userID int) (models.CalorieSettings, error) {
    log.Printf("[CalorieSettings] Търсене на настройки за потребител: %d", userID)
    
    var settings models.CalorieSettings
    
    err := db.QueryRow(`
        SELECT gender, age, activity_level, goal, updated_at 
        FROM calorie_settings 
        WHERE user_id = ?`, userID).Scan(
        &settings.Gender, &settings.Age, &settings.ActivityLevel, 
        &settings.Goal, &settings.UpdatedAt)

    if err != nil {
        log.Printf("[ERROR] Грешка при четене на настройки: %v", err)
        return settings, err
    }

    settings.UserID = userID
    return settings, nil
}

func getCalorieSettings(c *gin.Context) {
    userID := getUserID(c)
    
    settings, err := getCalorieSettingsInternal(userID)
	fmt.Println(settings)	
    if err == sql.ErrNoRows {
        c.JSON(http.StatusNotFound, gin.H{"error": "Настройките не са намерени"})
        return
    } else if err != nil {
        log.Printf("[ERROR] Грешка при четене на настройки: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на настройките"})
        return
    }

    c.JSON(http.StatusOK, settings)
}


func updateCalorieSettings(c *gin.Context) {
	userID := getUserID(c)
	var settings models.CalorieSettings
	
	if err := c.BindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings.UserID = userID

	_, err := db.Exec(`
        INSERT INTO calorie_settings (user_id, gender, age, activity_level, goal)
        VALUES (?, ?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
        gender = VALUES(gender),
        age = VALUES(age),
        activity_level = VALUES(activity_level),
        goal = VALUES(goal)`,
        settings.UserID, settings.Gender, settings.Age, 
        settings.ActivityLevel, settings.Goal)

    if err != nil {
        log.Printf("Грешка при обновяване на настройки за калории: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при запазване на настройките"})
        return
    }

    c.JSON(http.StatusOK, settings)
}

func getCalorieCalculations(c *gin.Context) {
	userID := getUserID(c)
	
	// Взимаме настройките на потребителя
	var settings models.CalorieSettings
	err := db.QueryRow(`
        SELECT gender, age, activity_level, goal 
        FROM calorie_settings 
        WHERE user_id = ?`, userID).Scan(
        &settings.Gender, &settings.Age, &settings.ActivityLevel, &settings.Goal)
    
    if err == sql.ErrNoRows {
        c.JSON(http.StatusNotFound, gin.H{"error": "Първо трябва да настроите вашите калории"})
        return
    } else if err != nil {
        log.Printf("Грешка при четене на настройки за калории: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изчисляване на калориите"})
        return
    }
    
    // Взимаме последното тегло и височината на потребителя
    var weight, height float64
    err = db.QueryRow(`
        SELECT w.weight, u.height 
        FROM weight_records w 
        JOIN users u ON u.id = w.user_id
        WHERE w.user_id = ? 
        ORDER BY w.created_at DESC 
        LIMIT 1`, userID).Scan(&weight, &height)
    
    if err != nil {
        log.Printf("Грешка при четене на тегло и височина: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изчисляване на калориите"})
        return
    }
    
    // Изчисляваме BMR по формулата на Mifflin-St Jeor
    var bmr float64
    if settings.Gender == "male" {
        bmr = 10*weight + 6.25*height - 5*float64(settings.Age) + 5
    } else {
        bmr = 10*weight + 6.25*height - 5*float64(settings.Age) - 161
    }
    
    // Коефициенти за активност
    activityFactors := map[string]float64{
        "sedentary":    1.2,  // Заседнал начин на живот
        "light":        1.375,// Леки упражнения 1-3 пъти седмично
        "moderate":     1.55, // Умерени упражнения 3-5 пъти седмично
        "active":       1.725,// Активни упражнения 6-7 пъти седмично
        "very_active":  1.9,  // Много активни упражнения всеки ден
    }
    
    // Изчисляваме калориите за поддържане
    maintenanceKcal := bmr * activityFactors[settings.ActivityLevel]
    
    // Изчисляваме целевите калории според целта
    var targetKcal float64
    switch settings.Goal {
    case "lose":
        targetKcal = maintenanceKcal - 500 // Дефицит от 500 калории за отслабване
    case "gain":
        targetKcal = maintenanceKcal + 500 // Излишък от 500 калории за качване
    default:
        targetKcal = maintenanceKcal
    }
    
    calculations := models.CalorieCalculation{
        BMR:             bmr,
        MaintenanceKcal: maintenanceKcal,
        TargetKcal:      targetKcal,
    }
    
    c.JSON(http.StatusOK, calculations)
}

func getDailyCalorieLog(c *gin.Context) {
	userID := getUserID(c)
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// Парсваме датата
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Грешка при парсване на дата: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Невалиден формат на дата"})
		return
	}
	
	// Взимаме или създаваме лог за деня
	var dailyLog models.DailyCalorieLog
	err = db.QueryRow(`
		SELECT id, food_kcal, activity_kcal, notes 
		FROM daily_calorie_logs 
		WHERE user_id = ? AND date = ?`,
		userID, date).Scan(&dailyLog.ID, &dailyLog.FoodKcal, &dailyLog.ActivityKcal, &dailyLog.Notes)
	
	log.Printf("Daily log: %+v", dailyLog, "SELECT id, food_kcal, activity_kcal, notes FROM daily_calorie_logs WHERE user_id = ? AND date = ?", userID, date)

	if err == sql.ErrNoRows {
		// Създаваме нов лог за деня
		result, err := db.Exec(`
			INSERT INTO daily_calorie_logs (user_id, date) 
			VALUES (?, ?)`,
			userID, date)
		
		if err != nil {
			log.Printf("Грешка при създаване на дневен ло3г: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при създаване на дневен лог3"})
			return
		}
		
		logID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Грешка при взимане на ID на новия лог: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при създаване на дневен лог4"})
			return
		}
		
		dailyLog = models.DailyCalorieLog{
			ID:     int(logID),
			UserID: userID,
			Date:   date,
		}
	} else if err != nil {
		log.Printf("Грешка при четене на дневен лог: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на дневния лог"})
		return
	}
	
	// Взимаме всички записи за храна за деня
	rows, err := db.Query(`
		SELECT id, name, calories, protein, carbs, fat, time, meal_type_id, notes
		FROM food_entries 
		WHERE user_id = ? AND log_id = ?
		ORDER BY time`,
		userID, dailyLog.ID)
	
	if err != nil {
		log.Printf("Грешка при четене на записи за храна: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на записите за храна"})
		return
	}
	defer rows.Close()
	
	var foodEntries []models.FoodEntry
	for rows.Next() {
		var entry models.FoodEntry
		err := rows.Scan(&entry.ID, &entry.Name, &entry.Calories,
			&entry.Protein, &entry.Carbs, &entry.Fat, &entry.Time, &entry.MealTypeID, &entry.Notes)
		if err != nil {
			continue
		}
		foodEntries = append(foodEntries, entry)
	}
	
	// Взимаме всички записи за активности за деня
	rows, err = db.Query(`
		SELECT id, type, duration, calories, time 
		FROM activity_entries 
		WHERE user_id = ? AND log_id = ?
		ORDER BY time`,
		userID, dailyLog.ID)
	
	if err != nil {
		log.Printf("Грешка при четене на записи за активности: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на записите за активности"})
		return
	}
	defer rows.Close()
	
	var activityEntries []models.ActivityEntry
	for rows.Next() {
		var entry models.ActivityEntry
		err := rows.Scan(&entry.ID, &entry.Type, &entry.Duration, &entry.Calories, &entry.Time)
		if err != nil {
			continue
		}
		activityEntries = append(activityEntries, entry)
	}

	if foodEntries == nil {
		foodEntries = []models.FoodEntry{}
	}
	if activityEntries == nil {
		activityEntries = []models.ActivityEntry{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"log": dailyLog,
		"foodEntries": foodEntries,
		"activityEntries": activityEntries,
	})
}

func addFoodEntry(c *gin.Context) {
	userID := getUserID(c)
	log.Printf("[FoodEntry] Добавяне на запис за потребител: %d", userID)
	
	var entry models.FoodEntry
	if err := c.BindJSON(&entry); err != nil {
		log.Printf("[ERROR] Невалидни данни: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	entry.UserID = userID
	// entry.Time = models.Now()
	entry.Time = time.Now()
	
	// Проверяваме дали съществува лог за деня
	var logID int
	query := `
		SELECT id FROM daily_calorie_logs 
		WHERE user_id = ? AND date = DATE(?)`

	err := db.QueryRow(query, userID, entry.Time).Scan(&logID)
	log.Printf("Log ID: %d", logID)
	if err == sql.ErrNoRows {
		// Създаваме нов лог за деня
		result, err := db.Exec(`
			INSERT INTO daily_calorie_logs (user_id, date, food_kcal) 
			VALUES (?, DATE(CURRENT_TIMESTAMP), ?)`,
			userID, entry.Calories)
		
		if err != nil {
			log.Printf("Грешка при създаване на дневен лог1: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при създаване на дневен лог1"})
			return
		}
		
		newLogID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Грешка при взимане на ID на новия лог: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при създаване на дневен лог2"})
			return
		}
		logID = int(newLogID)
	} else if err != nil {
		log.Printf("Грешка при проверка на дневен лог: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при проверка на дневния лог"})
		return
	} else {
		// Обновяваме общите калории за деня
		log.Printf("Обновяване на калориите за деня: %d", logID, entry.Calories)
		_, err = db.Exec(`
			UPDATE daily_calorie_logs 
			SET food_kcal = food_kcal + ? 
			WHERE id = ?`,
			entry.Calories, logID)
		
		if err != nil {
			log.Printf("Грешка при обновяване на калориите за деня: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при обновяване на калориите"})
			return
		}
	}
	
	entry.LogID = logID
	
	// Добавяме записа за храната
	result, err := db.Exec(`
		INSERT INTO food_entries (user_id, log_id, name, calories, protein, carbs, fat, time, meal_type_id, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.UserID, entry.LogID, entry.Name, entry.Calories,
		entry.Protein, entry.Carbs, entry.Fat, entry.Time, entry.MealTypeID, entry.Notes)
	
	if err != nil {
		log.Printf("[ERROR] Грешка при добавяне на запис: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при добавяне на записа"})
		return
	}
	
	newID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Грешка при взимане на ID на новия запис: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при добавяне на записа"})
		return
	}
	entry.ID = int(newID)
	
	c.JSON(http.StatusOK, entry)
}

func addActivityEntry(c *gin.Context) {
	userID := getUserID(c)
	var entry models.ActivityEntry
	
	if err := c.BindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	entry.UserID = userID
	
	// Проверяваме дали съществува лог за деня
	var logID int
	err := db.QueryRow(`
		SELECT id FROM daily_calorie_logs 
		WHERE user_id = ? AND date = DATE(?)`,
		userID, entry.Time).Scan(&logID)
	
	if err == sql.ErrNoRows {
		// Създаваме нов лог за деня
		result, err := db.Exec(`
			INSERT INTO daily_calorie_logs (user_id, date, activity_kcal) 
			VALUES (?, DATE(?), ?)`,
			userID, entry.Time, entry.Calories)
		
		if err != nil {
			log.Printf("Грешка при създаване на дневен лог: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при създаване на дневен лог"})
			return
		}
		
		newLogID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Грешка при взимане на ID на новия лог: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при създаване на дневен лог"})
			return
		}
		logID = int(newLogID)
	} else if err != nil {
		log.Printf("Грешка при проверка на дневен лог: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при проверка на дневния лог"})
		return
	} else {
		// Обновяваме общите изгорени калории за деня
		_, err = db.Exec(`
			UPDATE daily_calorie_logs 
			SET activity_kcal = activity_kcal + ? 
			WHERE id = ?`,
			entry.Calories, logID)
		
		if err != nil {
			log.Printf("Грешка при обновяване на калориите за деня: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при обновяване на калориите"})
			return
		}
	}
	
	entry.LogID = logID
	
	// Добавяме записа за активността
	result, err := db.Exec(`
		INSERT INTO activity_entries (user_id, log_id, type, duration, calories, time)
		VALUES (?, ?, ?, ?, ?, ?)`,
		entry.UserID, entry.LogID, entry.Type, entry.Duration, entry.Calories, entry.Time)
	
	if err != nil {
		log.Printf("Грешка при добавяне на запис за активност: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при добавяне на записа"})
		return
	}
	
	newID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Грешка при взимане на ID на новия запис: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при добавяне на записа"})
		return
	}
	entry.ID = int(newID)
	
	c.JSON(http.StatusOK, entry)
}

func getCalorieStats(c *gin.Context) {
	currentDate := time.Now()
	userID := getUserID(c)
	
	settings, err := getCalorieSettingsInternal(userID)
	if err != nil {
		log.Printf("[ERROR] Грешка при четене на настройки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на настройките"})
		return
	}
	
	startDate := c.DefaultQuery("startDate", currentDate.AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("endDate", currentDate.Format("2006-01-02"))
	
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Трябва да посочите начална и крайна дата"})
		return
	}
	
	// Взимаме дневните логове за периода
	rows, err := db.Query(`
		SELECT date, food_kcal, activity_kcal 
		FROM daily_calorie_logs 
		WHERE user_id = ? AND date BETWEEN ? AND ?
		ORDER BY date`,
		userID, startDate, endDate)
	
	if err != nil {
		log.Printf("Грешка при четене на статистика: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при четене на статистиката"})
		return
	}
	defer rows.Close()
	
	var stats []struct {
		Date         string  `json:"date"`
		FoodKcal     float64 `json:"foodKcal"`
		ActivityKcal float64 `json:"activityKcal"`
		NetKcal      float64 `json:"netKcal"`
	}
	
	for rows.Next() {
		var stat struct {
			Date         string  `json:"date"`
			FoodKcal     float64 `json:"foodKcal"`
			ActivityKcal float64 `json:"activityKcal"`
			NetKcal      float64 `json:"netKcal"`
		}
		err := rows.Scan(&stat.Date, &stat.FoodKcal, &stat.ActivityKcal)
		if err != nil {
			continue
		}
		log.Printf("[Calories] Статистика: %v", stat)
		stat.NetKcal = stat.FoodKcal - stat.ActivityKcal
		stats = append(stats, stat)
	}
	
	// Изчисляваме средните стойности
	var totalFood, totalActivity float64
	var daysCount int
	for _, stat := range stats {
		totalFood += stat.FoodKcal
		totalActivity += stat.ActivityKcal
		daysCount++
	}
	
	var averages struct {
		FoodKcal     float64 `json:"foodKcal"`
		ActivityKcal float64 `json:"activityKcal"`
		NetKcal      float64 `json:"netKcal"`
	}
	
	if daysCount > 0 {
		averages.FoodKcal = totalFood / float64(daysCount)
		averages.ActivityKcal = totalActivity / float64(daysCount)
		averages.NetKcal = averages.FoodKcal - averages.ActivityKcal
	}

	//===============================================

	log.Printf("[Calories] Изчисляване на нужни калории за потребител: %d", userID)

    // Взимаме данните на потребителя
    var (
        weight, height float64
        age           int
        gender        string
        activityLevel string
		goal          string
    )

    err = db.QueryRow(`
        SELECT w.weight, u.height, u.age, cs.gender, cs.activity_level, cs.goal
        FROM weight_records w 
        JOIN users u ON u.id = w.user_id
        JOIN calorie_settings cs ON cs.user_id = u.id
        WHERE w.user_id = ? 
        ORDER BY w.created_at DESC 
        LIMIT 1`, userID).Scan(&weight, &height, &age, &gender, &activityLevel, &goal)

    if err != nil {
        log.Printf("[ERROR] Грешка при четене на данни: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изчисляване"})
        return
    }

    // Изчисляваме BMR
    bmr := calculateBMR(weight, height, age, gender)
    
    // Изчисляваме дневните нужди според активността
    maintenanceCalories := bmr * getActivityMultiplier(activityLevel)
    
    // Изчисляваме калории за отслабване и качване
    // За отслабване: дефицит от 20% от поддържащите калории
    weightLossCalories := math.Round(maintenanceCalories * 0.8)
    
    // За качване: излишък от 20% от поддържащите калории
    weightGainCalories := math.Round(maintenanceCalories * 1.2)

	targetCalories := float64(0)
	if goal == "maintain" {
		targetCalories = maintenanceCalories
	} else if goal == "lose" {
		targetCalories = weightLossCalories
	} else if goal == "gain" {
		targetCalories = weightGainCalories
	}

    // Изчисляваме дневния прием за всеки ден
    var totalDailyIntake float64
	var totalCurrentDayIntake float64
    var daysWithIntake int
    currentDateStr := currentDate.Format("2006-01-02")


    for _, stat := range stats {
        // Дневен прием = приети калории - изгорени калории
        dailyIntake := stat.FoodKcal - stat.ActivityKcal
        
        // Добавяме към общата сума само ако има данни за деня
        if stat.FoodKcal > 0 || stat.ActivityKcal > 0 {
			totalDailyIntake += dailyIntake
			daysWithIntake++

			if stat.Date[:10] == currentDateStr[:10] {
				totalCurrentDayIntake = stat.FoodKcal
			}
        }
    }
    
    // Изчисляваме средния дневен прием
    averageDailyIntake := float64(0)
    if daysWithIntake > 0 {
        averageDailyIntake = totalDailyIntake / float64(daysWithIntake)
    }
    
    // Изчисляваме разликата от целта
    calorieDeficit := float64(0)
    if targetCalories > 0 {
        calorieDeficit = averageDailyIntake - targetCalories
    }
    
    // Добавяме анализ на приема
    intakeAnalysis := struct {
        AverageDailyIntake float64 `json:"averageDailyIntake"`
		TotalDailyIntake   float64 `json:"totalDailyIntake"`
		TotalCurrentDayIntake float64 `json:"totalCurrentDayIntake"`
        TargetCalories    float64 `json:"targetCalories"`
        CalorieDeficit    float64 `json:"calorieDeficit"`
        Status            string  `json:"status"`
    }{
        AverageDailyIntake: math.Round(averageDailyIntake),
		TotalDailyIntake:   math.Round(totalDailyIntake),
		TotalCurrentDayIntake: math.Round(totalCurrentDayIntake),
        TargetCalories:    math.Round(targetCalories),
        CalorieDeficit:    math.Round(calorieDeficit),
        Status:            getCalorieStatus(calorieDeficit, goal),
    }

    c.JSON(http.StatusOK, gin.H{
        "settings": settings,
        "targetCalories": targetCalories,
        "daily": stats,
        "averages": averages,
        "currentWeight": weight,
        "bmr": math.Round(bmr),
        "dailyNeeds": map[string]float64{
            "maintenance": math.Round(maintenanceCalories),
            "weightLoss": math.Round(weightLossCalories),
            "weightGain": math.Round(weightGainCalories),
        },
        "macronutrients": map[string]map[string]float64{
            "maintenance": calculateMacros(maintenanceCalories),
            "weightLoss": calculateMacros(weightLossCalories),
            "weightGain": calculateMacros(weightGainCalories),
        },
        "intakeAnalysis": intakeAnalysis,
    })
}

// Помощна функция за определяне на статуса
func getCalorieStatus(deficit float64, goal string) string {
    switch goal {
    case "lose":
        if deficit >= 0 {
            return "над целта"
        }
        return "под целта"
    case "gain":
        if deficit >= 0 {
            return "над целта"
        }
        return "под целта"
    case "maintain":
        if math.Abs(deficit) <= 100 { // Допускаме отклонение от 100 калории
            return "в целта"
        } else if deficit > 0 {
            return "над целта"
        }
        return "под целта"
    default:
        return "неопределен"
    }
}

func deleteFoodEntry(c *gin.Context) {
	userID := getUserID(c)
	entryID := c.Param("id")

	// Първо проверяваме дали записът принадлежи на този потребител
	var logID int
	var calories float64
	err := db.QueryRow(`
		SELECT log_id, calories 
		FROM food_entries 
		WHERE id = ? AND user_id = ?`,
		entryID, userID).Scan(&logID, &calories)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Записът не е намерен"})
		return
	} else if err != nil {
		log.Printf("Грешка при проверка на записа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	// Изтриваме записа
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Грешка при започване на транзакция: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	// Изтриваме записа за храната
	_, err = tx.Exec("DELETE FROM food_entries WHERE id = ?", entryID)
	if err != nil {
		tx.Rollback()
		log.Printf("Грешка при изтриване на запис: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	// Обновяваме общите калории в дневния лог
	_, err = tx.Exec(`
		UPDATE daily_calorie_logs 
		SET food_kcal = food_kcal - ? 
		WHERE id = ?`,
		calories, logID)
	if err != nil {
		tx.Rollback()
		log.Printf("Грешка при обновяване на дневния лог: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при обновяване на дневния лог"})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Грешка при завършване на транзакция: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Записът е изтрит успешно"})
}

func deleteActivityEntry(c *gin.Context) {
	userID := getUserID(c)
	entryID := c.Param("id")

	// Първо проверяваме дали записът принадлежи на този потребител
	var logID int
	var calories float64
	err := db.QueryRow(`
		SELECT log_id, calories 
		FROM activity_entries 
		WHERE id = ? AND user_id = ?`,
		entryID, userID).Scan(&logID, &calories)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Записът не е намерен"})
		return
	} else if err != nil {
		log.Printf("Грешка при проверка на записа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	// Изтриваме записа
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Грешка при започване на транзакция: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	// Изтриваме записа за активността
	_, err = tx.Exec("DELETE FROM activity_entries WHERE id = ?", entryID)
	if err != nil {
		tx.Rollback()
		log.Printf("Грешка при изтриване на запис: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	// Обновяваме общите изгорени калории в дневния лог
	_, err = tx.Exec(`
		UPDATE daily_calorie_logs 
		SET activity_kcal = activity_kcal - ? 
		WHERE id = ?`,
		calories, logID)
	if err != nil {
		tx.Rollback()
		log.Printf("Грешка при обновяване на дневния лог: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при обновяване на дневния лог"})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Грешка при завършване на транзакция: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изтриване на записа"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Записът е изтрит успешно"})
}

func getMealTypes(c *gin.Context) {
    log.Printf("[MealTypes] Започва зареждане на типове храна")
    
    var mealTypes []models.MealType

	rows, err := db.Query("SELECT id, name, created_at, updated_at FROM meal_types")
	if err != nil {
		log.Printf("Грешка при заявка за типове храна: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при взимане на типове храна"})
		return
	}
	defer rows.Close()

    for rows.Next() {
        var mealType models.MealType
        var name []byte // Временна променлива за името
        
        // Сканираме името като []byte вместо string
        err := rows.Scan(&mealType.ID, &name, &mealType.CreatedAt, &mealType.UpdatedAt)
        if err != nil {
            log.Printf("[ERROR] Грешка при сканиране на тип храна: %v", err)
            continue
        }
        
        // Конвертираме bytes към UTF-8 string
        mealType.Name = string(name)
        
        mealTypes = append(mealTypes, mealType)
    }

    if mealTypes == nil {
        mealTypes = []models.MealType{}
    }

    // Задаваме изрично Content-Type хедъра
    c.Header("Content-Type", "application/json; charset=utf8")

    c.JSON(http.StatusOK, mealTypes)
}

// Калкулира базовия метаболизъм (BMR) по формулата на Mifflin-St Jeor
func calculateBMR(weight float64, height float64, age int, gender string) float64 {
    // BMR = (10 × weight) + (6.25 × height) - (5 × age) + s
    // където s е +5 за мъже и -161 за жени
    if gender == "male" {
        return 10*weight + 6.25*height - 5*float64(age) + 5
    }
    return 10*weight + 6.25*height - 5*float64(age) - 161
}

// Връща множител според нивото на активност
func getActivityMultiplier(activityLevel string) float64 {
    switch activityLevel {
    case "sedentary":     // Заседнал начин на живот
        return 1.2
    case "light":         // Леки упражнения 1-3 пъти седмично
        return 1.375
    case "moderate":      // Умерени упражнения 3-5 пъти седмично
        return 1.55
    case "active":        // Активни упражнения 6-7 пъти седмично
        return 1.725
    case "very_active":   // Много активни упражнения и физическа работа
        return 1.9
    default:
        return 1.2
    }
}

func calculateCalorieNeeds(c *gin.Context) {
    userID := getUserID(c)
    log.Printf("[Calories] Изчисляване на нужни калории за потребител: %d", userID)

    // Взимаме данните на потребителя
    var (
        weight, height float64
        age           int
        gender        string
        activityLevel string
    )

    err := db.QueryRow(`
        SELECT w.weight, u.height, u.age, cs.gender, cs.activity_level
        FROM weight_records w 
        JOIN users u ON u.id = w.user_id
        JOIN calorie_settings cs ON cs.user_id = u.id
        WHERE w.user_id = ? 
        ORDER BY w.created_at DESC 
        LIMIT 1`, userID).Scan(&weight, &height, &age, &gender, &activityLevel)

    if err != nil {
        log.Printf("[ERROR] Грешка при четене на данни: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Грешка при изчисляване"})
        return
    }

    // Изчисляваме BMR
    bmr := calculateBMR(weight, height, age, gender)
    
    // Изчисляваме дневните нужди според активността
    maintenanceCalories := bmr * getActivityMultiplier(activityLevel)
    
    // Изчисляваме калории за отслабване и качване
    // За отслабване: дефицит от 20% от поддържащите калории
    weightLossCalories := maintenanceCalories * 0.8
    
    // За качване: излишък от 20% от поддържащите калории
    weightGainCalories := maintenanceCalories * 1.2

    response := gin.H{
        "currentWeight": weight,
        "bmr": math.Round(bmr),
        "dailyNeeds": map[string]float64{
            "maintenance": math.Round(maintenanceCalories),
            "weightLoss": math.Round(weightLossCalories),
            "weightGain": math.Round(weightGainCalories),
        },
        "macronutrients": map[string]map[string]float64{
            "maintenance": calculateMacros(maintenanceCalories),
            "weightLoss": calculateMacros(weightLossCalories),
            "weightGain": calculateMacros(weightGainCalories),
        },
    }

    c.JSON(http.StatusOK, response)
}

// Изчислява препоръчителното разпределение на макронутриентите
func calculateMacros(calories float64) map[string]float64 {
    // Протеин: 30%, Въглехидрати: 40%, Мазнини: 30%
    return map[string]float64{
        "protein": math.Round(calories * 0.30 / 4),  // 4 калории на грам протеин
        "carbs":   math.Round(calories * 0.40 / 4),  // 4 калории на грам въглехидрати
        "fat":     math.Round(calories * 0.30 / 9),  // 9 калории на грам мазнини
    }
}