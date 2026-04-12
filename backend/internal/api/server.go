package api

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rueditor/backend/internal/models"
	"github.com/rueditor/backend/internal/services"
	"github.com/rueditor/backend/internal/websocket"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	engine       *gin.Engine
	latexService *services.LatexService
	wsManager    *websocket.Manager
	port         string
	jwtSecret    []byte
	documents    map[uuid.UUID]*models.Document
	users        map[uuid.UUID]*models.User
}

func NewServer(latexService *services.LatexService, wsManager *websocket.Manager) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	engine.Use(corsMiddleware())
	engine.Use(gin.Recovery())

	srv := &Server{
		engine:       engine,
		latexService: latexService,
		wsManager:    wsManager,
		port:         getEnv("PORT", "8080"),
		jwtSecret:    []byte(getEnv("JWT_SECRET", "supersecretkey123")),
		documents:    make(map[uuid.UUID]*models.Document),
		users:        make(map[uuid.UUID]*models.User),
	}

	srv.setupRoutes()
	return srv
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func (s *Server) setupRoutes() {
	api := s.engine.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)

		auth := api.Group("/auth")
		{
			auth.POST("/register", s.register)
			auth.POST("/login", s.login)
		}

		docs := api.Group("/documents")
		docs.Use(s.authMiddleware())
		{
			docs.GET("", s.listDocuments)
			docs.POST("", s.createDocument)
			docs.GET("/:id", s.getDocument)
			docs.PUT("/:id", s.updateDocument)
			docs.DELETE("/:id", s.deleteDocument)
			docs.POST("/:id/compile", s.compileDocument)
			docs.POST("/:id/collaborate", s.inviteCollaborator)
		}
	}
}

func (s *Server) Run() error {
	return s.engine.Run(":" + s.port)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "rueditor-api",
		"time":    time.Now().UTC(),
	})
}

func (s *Server) register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.users[user.ID] = user

	token, err := s.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (s *Server) login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user *models.User
	for _, u := range s.users {
		if u.Email == req.Email {
			user = u
			break
		}
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (s *Server) listDocuments(c *gin.Context) {
	userID := s.getUserIDFromContext(c)

	var userDocs []*models.Document
	for _, doc := range s.documents {
		if doc.UserID == userID || containsUUID(doc.Collaborators, userID) {
			userDocs = append(userDocs, doc)
		}
	}

	c.JSON(http.StatusOK, gin.H{"documents": userDocs})
}

func (s *Server) createDocument(c *gin.Context) {
	userID := s.getUserIDFromContext(c)

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc := &models.Document{
		ID:        uuid.New(),
		Title:     req.Title,
		Content:   req.Content,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if doc.Content == "" {
		doc.Content = getDefaultLatexTemplate()
	}

	s.documents[doc.ID] = doc
	c.JSON(http.StatusCreated, doc)
}

func (s *Server) getDocument(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, exists := s.documents[docID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if doc.UserID != userID && !containsUUID(doc.Collaborators, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (s *Server) updateDocument(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, exists := s.documents[docID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if doc.UserID != userID && !containsUUID(doc.Collaborators, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Content != "" {
		doc.Content = req.Content
	}
	doc.UpdatedAt = time.Now()

	s.wsManager.BroadcastToDocument(docID, websocket.Message{
		Type:       "document_update",
		DocumentID: docID,
		UserID:     userID,
		Content:    req.Content,
		Timestamp:  time.Now(),
	})

	c.JSON(http.StatusOK, doc)
}

func (s *Server) deleteDocument(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, exists := s.documents[docID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if doc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	delete(s.documents, docID)
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

func (s *Server) compileDocument(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc := s.documents[docID]
	content := req.Content
	if content == "" {
		if doc != nil {
			content = doc.Content
		}
	}

	result, err := s.latexService.Compile(content, docID.String())
	if err != nil {
		c.JSON(http.StatusOK, models.CompileResponse{
			Success: false,
			Error:   err.Error(),
			Log:     result.Log,
		})
		return
	}

	c.JSON(http.StatusOK, models.CompileResponse{
		Success: true,
		PDF:     result.PDF,
		Log:     result.Log,
	})
}

func (s *Server) inviteCollaborator(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, exists := s.documents[docID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if doc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can invite collaborators"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var collaboratorID uuid.UUID
	for _, u := range s.users {
		if u.Email == req.Email {
			collaboratorID = u.ID
			break
		}
	}

	if collaboratorID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	doc.Collaborators = append(doc.Collaborators, collaboratorID)
	c.JSON(http.StatusOK, doc)
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func (s *Server) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Server) getUserIDFromContext(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func containsUUID(slice []uuid.UUID, item uuid.UUID) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func getDefaultLatexTemplate() string {
	return `\documentclass[11pt]{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{geometry}
\geometry{a4paper, margin=1in}
\usepackage{hyperref}
\usepackage{graphicx}
\usepackage{parskip}

\title{Your Resume Title}
\author{Your Name}
\date{\today}

\begin{document}

\maketitle

\section{Education}
\begin{itemize}
    \item Degree in Field, University Name, Year
\end{itemize}

\section{Experience}
\begin{itemize}
    \item \textbf{Job Title} - Company Name
    \begin{itemize}
        \item Description of responsibilities and achievements
    \end{itemize}
\end{itemize}

\section{Skills}
\begin{itemize}
    \item Programming Languages: Python, JavaScript, Go
    \item Tools: Git, Docker, Linux
\end{itemize}

\end{document}
`
}
