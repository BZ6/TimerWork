package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"-" db:"password"`
}

type WorkWeek struct {
	ID            int       `json:"id" db:"id"`
	UserID        int       `json:"user_id" db:"user_id"`
	WeekStart     time.Time `json:"week_start" db:"week_start"`
	WeekEnd       *time.Time `json:"week_end" db:"week_end"`
	LastUpdateTime time.Time `json:"last_update_time" db:"last_update_time"`
	Status        string    `json:"status" db:"status"` // "running", "paused", "stopped"
	PauseStart    *time.Time `json:"pause_start" db:"pause_start"`
	TotalPauseTime int64     `json:"total_pause_time" db:"total_pause_time"` // в секундах
	WeekGoalMinutes int      `json:"week_goal_minutes" db:"week_goal_minutes"` // цель недели в минутах
}

type WorkWeekResponse struct {
	WorkWeek
	ElapsedTime int64 `json:"elapsed_time"` // в секундах
}

type WorkWeekHistoryItem struct {
	ID                int        `json:"id" db:"id"`
	WeekNumber        int        `json:"week_number"`
	StartedAt         *time.Time `json:"started_at" db:"week_start"`
	EndedAt           *time.Time `json:"ended_at" db:"week_end"`
	TotalWorkMinutes  int        `json:"total_work_minutes"`
	WeekGoalMinutes   int        `json:"week_goal_minutes"`
	IsPaused          bool       `json:"is_paused"`
	Status            string     `json:"status" db:"status"`
}

type Database struct {
	db *sql.DB
}

type Server struct {
	db        *Database
	jwtSecret string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type StartWeekRequest struct {
	GoalMinutes int `json:"goal_minutes"`
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func main() {
	// Подключение к базе данных
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "admin")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "timerwork")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key")

	dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Проверка соединения с повторными попытками
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... attempt %d/30", i+1)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		log.Fatal("Failed to ping database after 30 attempts:", err)
	}

	database := &Database{db: db}
	server := &Server{db: database, jwtSecret: jwtSecret}

	// Создание таблиц
	if err := database.createTables(); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Настройка роутера
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Публичные маршруты
	r.POST("/api/register", server.register)
	r.POST("/api/login", server.login)

	// Защищённые маршруты
	protected := r.Group("/api")
	protected.Use(server.authMiddleware())
	{
		protected.GET("/workweek", server.getWorkWeek)
		protected.GET("/workweek/history", server.getWorkWeekHistory)
		protected.POST("/workweek/start", server.startWeek)
		protected.POST("/workweek/end", server.endWeek)
		protected.POST("/workweek/pause", server.pauseTimer)
		protected.POST("/workweek/resume", server.resumeTimer)
		protected.GET("/workweek/current-time", server.getCurrentWorkTime)
	}

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (db *Database) createTables() error {
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)`

	workWeekTable := `
	CREATE TABLE IF NOT EXISTS work_weeks (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		week_start TIMESTAMP NOT NULL,
		week_end TIMESTAMP,
		last_update_time TIMESTAMP NOT NULL DEFAULT NOW(),
		status VARCHAR(20) NOT NULL DEFAULT 'stopped',
		pause_start TIMESTAMP,
		total_pause_time INTEGER DEFAULT 0,
		week_goal_minutes INTEGER DEFAULT 2400,
		created_at TIMESTAMP DEFAULT NOW()
	)`

	if _, err := db.db.Exec(userTable); err != nil {
		return err
	}

	if _, err := db.db.Exec(workWeekTable); err != nil {
		return err
	}

	// Миграция: добавляем колонку week_goal_minutes, если её нет
	_, err := db.db.Exec(`
		ALTER TABLE work_weeks 
		ADD COLUMN IF NOT EXISTS week_goal_minutes INTEGER DEFAULT 2400
	`)
	if err != nil {
		log.Printf("Warning: Could not add week_goal_minutes column: %v", err)
	}

	return nil
}

func (s *Server) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создание пользователя
	var userID int
	err = s.db.db.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		req.Username, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user_id": userID})
}

func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Поиск пользователя
	var user User
	err := s.db.db.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Создание JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokenString,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Убираем "Bearer " если есть
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func (s *Server) getWorkWeek(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var workWeek WorkWeek
	err := s.db.db.QueryRow(`
		SELECT id, user_id, week_start, week_end, last_update_time, status, pause_start, total_pause_time, week_goal_minutes
		FROM work_weeks 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, userID).Scan(
		&workWeek.ID,
		&workWeek.UserID,
		&workWeek.WeekStart,
		&workWeek.WeekEnd,
		&workWeek.LastUpdateTime,
		&workWeek.Status,
		&workWeek.PauseStart,
		&workWeek.TotalPauseTime,
		&workWeek.WeekGoalMinutes,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"work_week": nil})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	response := WorkWeekResponse{
		WorkWeek:    workWeek,
		ElapsedTime: s.calculateElapsedTime(workWeek),
	}

	c.JSON(http.StatusOK, gin.H{"work_week": response})
}

func (s *Server) startWeek(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req StartWeekRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Устанавливаем значение по умолчанию, если цель не указана
	if req.GoalMinutes <= 0 {
		req.GoalMinutes = 2400 // 40 часов по умолчанию
	}

	now := time.Now()
	
	// Проверяем, есть ли активная неделя
	var existingID int
	err := s.db.db.QueryRow(
		"SELECT id FROM work_weeks WHERE user_id = $1 AND week_end IS NULL ORDER BY created_at DESC LIMIT 1",
		userID,
	).Scan(&existingID)

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Week already started"})
		return
	}

	// Создаём новую рабочую неделю
	var workWeekID int
	err = s.db.db.QueryRow(`
		INSERT INTO work_weeks (user_id, week_start, last_update_time, status, week_goal_minutes) 
		VALUES ($1, $2, $2, 'running', $3) 
		RETURNING id
	`, userID, now, req.GoalMinutes).Scan(&workWeekID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start week"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Week started", "work_week_id": workWeekID})
}

func (s *Server) endWeek(c *gin.Context) {
	userID, _ := c.Get("user_id")

	now := time.Now()

	// Обновляем время окончания недели
	result, err := s.db.db.Exec(`
		UPDATE work_weeks 
		SET week_end = $1, last_update_time = $1, status = 'stopped'
		WHERE user_id = $2 AND week_end IS NULL
	`, now, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end week"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active week found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Week ended"})
}

func (s *Server) pauseTimer(c *gin.Context) {
	userID, _ := c.Get("user_id")

	now := time.Now()

	result, err := s.db.db.Exec(`
		UPDATE work_weeks 
		SET status = 'paused', pause_start = $1, last_update_time = $1
		WHERE user_id = $2 AND week_end IS NULL AND status = 'running'
	`, now, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause timer"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No running timer found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Timer paused"})
}

func (s *Server) resumeTimer(c *gin.Context) {
	userID, _ := c.Get("user_id")

	now := time.Now()

	// Получаем время начала паузы
	var pauseStart *time.Time
	var totalPauseTime int64
	err := s.db.db.QueryRow(`
		SELECT pause_start, total_pause_time 
		FROM work_weeks 
		WHERE user_id = $1 AND week_end IS NULL AND status = 'paused'
	`, userID).Scan(&pauseStart, &totalPauseTime)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No paused timer found"})
		return
	}

	// Вычисляем время паузы
	var pauseDuration int64
	if pauseStart != nil {
		pauseDuration = int64(now.Sub(*pauseStart).Seconds())
	}

	// Обновляем запись
	result, err := s.db.db.Exec(`
		UPDATE work_weeks 
		SET status = 'running', pause_start = NULL, last_update_time = $1, total_pause_time = $2
		WHERE user_id = $3 AND week_end IS NULL AND status = 'paused'
	`, now, totalPauseTime+pauseDuration, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume timer"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No paused timer found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Timer resumed"})
}

func (s *Server) getCurrentWorkTime(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var workWeek WorkWeek
	err := s.db.db.QueryRow(`
		SELECT id, user_id, week_start, week_end, last_update_time, status, pause_start, total_pause_time
		FROM work_weeks 
		WHERE user_id = $1 AND week_end IS NULL
		ORDER BY created_at DESC 
		LIMIT 1
	`, userID).Scan(
		&workWeek.ID,
		&workWeek.UserID,
		&workWeek.WeekStart,
		&workWeek.WeekEnd,
		&workWeek.LastUpdateTime,
		&workWeek.Status,
		&workWeek.PauseStart,
		&workWeek.TotalPauseTime,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"elapsed_time": 0, "status": "stopped"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Обновляем last_update_time если таймер запущен
	if workWeek.Status == "running" {
		now := time.Now()
		_, err = s.db.db.Exec(
			"UPDATE work_weeks SET last_update_time = $1 WHERE id = $2",
			now, workWeek.ID,
		)
		if err == nil {
			workWeek.LastUpdateTime = now
		}
	}

	elapsedTime := s.calculateElapsedTime(workWeek)

	c.JSON(http.StatusOK, gin.H{
		"elapsed_time": elapsedTime,
		"status":       workWeek.Status,
		"week_start":   workWeek.WeekStart,
		"week_end":     workWeek.WeekEnd,
	})
}

func (s *Server) calculateElapsedTime(workWeek WorkWeek) int64 {
	if workWeek.Status == "stopped" && workWeek.WeekEnd != nil {
		// Неделя закончена
		totalTime := int64(workWeek.WeekEnd.Sub(workWeek.WeekStart).Seconds())
		return totalTime - workWeek.TotalPauseTime
	}

	var currentTime time.Time
	var additionalPause int64

	if workWeek.Status == "running" {
		currentTime = time.Now()
	} else if workWeek.Status == "paused" {
		currentTime = workWeek.LastUpdateTime
		if workWeek.PauseStart != nil {
			// Добавляем текущую паузу
			additionalPause = int64(time.Now().Sub(*workWeek.PauseStart).Seconds())
		}
	} else {
		currentTime = workWeek.LastUpdateTime
	}

	totalTime := int64(currentTime.Sub(workWeek.WeekStart).Seconds())
	return totalTime - workWeek.TotalPauseTime - additionalPause
}

func (s *Server) getWorkWeekHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	query := `
		SELECT 
			id, 
			week_start, 
			week_end, 
			status,
			total_pause_time,
			week_goal_minutes
		FROM work_weeks 
		WHERE user_id = $1 
		ORDER BY week_start DESC
		LIMIT 50
	`

	rows, err := s.db.db.Query(query, userID)
	if err != nil {
		log.Printf("Error fetching work week history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения истории недель"})
		return
	}
	defer rows.Close()

	var history []WorkWeekHistoryItem
	weekNumber := 1

	for rows.Next() {
		var item WorkWeekHistoryItem
		var weekStart time.Time
		var weekEnd *time.Time
		var status string
		var totalPauseTime int64
		var weekGoalMinutes int

		err := rows.Scan(&item.ID, &weekStart, &weekEnd, &status, &totalPauseTime, &weekGoalMinutes)
		if err != nil {
			log.Printf("Error scanning work week row: %v", err)
			continue
		}

		item.StartedAt = &weekStart
		item.EndedAt = weekEnd
		item.Status = status
		item.WeekNumber = weekNumber
		item.WeekGoalMinutes = weekGoalMinutes
		item.IsPaused = (status == "paused")

		// Вычисляем продолжительность в минутах
		if weekEnd != nil {
			totalSeconds := int64(weekEnd.Sub(weekStart).Seconds()) - totalPauseTime
			item.TotalWorkMinutes = int(totalSeconds / 60)
		} else if status == "running" || status == "paused" {
			var currentTime time.Time
			if status == "running" {
				currentTime = time.Now()
			} else {
				currentTime = weekStart // fallback
			}
			totalSeconds := int64(currentTime.Sub(weekStart).Seconds()) - totalPauseTime
			item.TotalWorkMinutes = int(totalSeconds / 60)
		}

		history = append(history, item)
		weekNumber++
	}

	c.JSON(http.StatusOK, history)
}
