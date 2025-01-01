package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"weight-challenge/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {
	// Зареждане на .env файл
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Свързване с базата данни
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
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

	// Проверка на връзката
	if err := db.Ping(); err != nil {
		log.Fatal("Could not connect to database:", err)
	}

	log.Println("Successfully connected to database")
	defer db.Close()

	r := gin.Default()

	// Добавяме debug логове
	gin.SetMode(gin.DebugMode)
	r.Use(gin.Logger())

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

	// Защитени endpoints
	authorized := r.Group("/")
	authorized.Use(authMiddleware())
	{
		authorized.POST("/weight", addWeight)
		authorized.GET("/weight/stats", getWeightStats)
		authorized.DELETE("/weight/:id", deleteWeight)
		authorized.GET("/user/settings", getUserSettings)
		authorized.PUT("/user/settings", updateUserSettings)
		authorized.PUT("/user/password", changePassword)

		// Нови endpoints за социални функции
		authorized.GET("/users", getVisibleUsers)
		authorized.PUT("/user/visibility", updateVisibility)

		// Приятелства
		authorized.GET("/friends", getFriends)
		authorized.POST("/friends/request/:userId", sendFriendRequest)
		authorized.PUT("/friends/accept/:requestId", acceptFriendRequest)
		authorized.PUT("/friends/reject/:requestId", rejectFriendRequest)

		// Съревнования
		authorized.GET("/challenges", getChallenges)
		authorized.POST("/challenges", createChallenge)
		authorized.PUT("/challenges/:challengeId/accept", acceptChallenge)
		authorized.PUT("/challenges/:challengeId/reject", rejectChallenge)
		authorized.GET("/challenges/:challengeId/results", getChallengeResults)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://%s:%s", os.Getenv("API_URL"), port)
	r.Run(":" + port)
}

// func register(c *gin.Context) {
// 	var user models.User
// 	if err := c.BindJSON(&user); err != nil {
// 		log.Printf("Error binding JSON: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	log.Printf("Registration attempt for user: %s with password: %s", user.Username, user.Password)

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		log.Printf("Error hashing password: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
// 		return
// 	}

// 	log.Printf("Generated hashed password: %s", string(hashedPassword))

// 	result, err := db.Exec("INSERT INTO users (username, password, height) VALUES (?, ?, ?)",
// 		user.Username, string(hashedPassword), user.Height)
// 	if err != nil {
// 		log.Printf("Database error: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
// 		return
// 	}

// 	id, _ := result.LastInsertId()
// 	user.ID = int(id)
// 	user.Password = "" // Не връщаме паролата

// 	log.Printf("Successfully registered user: ID=%d, Username=%s", user.ID, user.Username)

// 	c.JSON(http.StatusOK, user)
// }

// func login(c *gin.Context) {
// 	var credentials struct {
// 		Username string `json:"username"`
// 		Password string `json:"password"`
// 	}

// 	if err := c.BindJSON(&credentials); err != nil {
// 		log.Printf("Error binding JSON: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	log.Printf("Login attempt for user: %s", credentials.Username)

// 	var user models.User
// 	var hashedPassword string
// 	err := db.QueryRow("SELECT id, username, password, height FROM users WHERE username = ?",
// 		credentials.Username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Height)

// 	if err != nil {
// 		log.Printf("Database error: %v", err)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	log.Printf("Found user: ID=%d, Username=%s", user.ID, user.Username)

// 	CheckPasswordHash(credentials.Password, hashedPassword)

// 	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password)); err != nil {
// 		log.Printf("Password comparison failed: %v", err)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	log.Printf("Login successful for user: %s", user.Username)

// 	c.JSON(http.StatusOK, gin.H{
// 		"user":  user,
// 		"token": "dummy-token",
// 	})
// }

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
        SELECT id, username, password, height 
        FROM users 
        WHERE username = ?`,
		credentials.Username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Height)

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
	token := fmt.Sprintf("user-%d", user.ID)

	log.Printf("Successful login for user: %s", credentials.Username)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
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
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// Извличаме ID-то от токена
		var userID int
		_, err := fmt.Sscanf(token, "user-%d", &userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Запазваме ID-то в контекста
		c.Set("userID", userID)
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
	userID := getUserID(c)

	var user models.User
	err := db.QueryRow(`
        SELECT id, username, first_name, last_name, age, height, gender, email, target_weight 
        FROM users 
        WHERE id = ?`, userID).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName,
		&user.Age, &user.Height, &user.Gender, &user.Email, &user.Target)

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
	friendID := c.Param("userId")

	_, err := db.Exec(`
        INSERT INTO friendships (requester_id, addressee_id) 
        VALUES (?, ?)`,
		userID, friendID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not send friend request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

func acceptFriendRequest(c *gin.Context) {
	userID := getUserID(c)
	requestID := c.Param("requestId")

	result, err := db.Exec(`
        UPDATE friendships 
        SET status = 'accepted' 
        WHERE id = ? AND addressee_id = ?`,
		requestID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not accept friend request"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

func rejectFriendRequest(c *gin.Context) {
	userID := getUserID(c)
	requestID := c.Param("requestId")

	result, err := db.Exec(`
        UPDATE friendships 
        SET status = 'rejected' 
        WHERE id = ? AND addressee_id = ?`,
		requestID, userID)

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
