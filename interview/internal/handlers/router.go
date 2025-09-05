package handlers

import (
	"github.com/gin-gonic/gin"
	"interview/internal/service"
)

func SetupRouter(
	vacancySvc service.VacancyService,
	resumeSvc service.ResumeService,
	interviewSvc service.InterviewService,
) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "interview"})
	})

	// Создаём handlers
	vacancyH := NewVacancyHandler(vacancySvc)
	resumeH := NewResumeHandler(resumeSvc)
	interviewH := NewHandler(interviewSvc)

	api := r.Group("/api")

	// Vacancy routes - ИСПОЛЬЗУЙ :id ВЕЗДЕ!
	vacancies := api.Group("/vacancies")
	{
		vacancies.GET("", vacancyH.GetAll)                       // GET /api/vacancies
		vacancies.POST("", vacancyH.Create)                      // POST /api/vacancies
		vacancies.GET("/:id", vacancyH.GetByID)                  // GET /api/vacancies/:id
		vacancies.PUT("/:id", vacancyH.Update)                   // PUT /api/vacancies/:id
		vacancies.DELETE("/:id", vacancyH.Delete)                // DELETE /api/vacancies/:id
		vacancies.GET("/:id/download", vacancyH.GetDownloadLink) // GET /api/vacancies/:id/download
		vacancies.GET("/:id/resumes", resumeH.GetByVacancy)      // ← ИСПРАВЛЕНО! Используй :id
	}

	// Resume routes
	resumes := api.Group("/resumes")
	{
		resumes.POST("", resumeH.Create)                      // POST /api/resumes
		resumes.GET("/:id", resumeH.GetByID)                  // GET /api/resumes/:id
		resumes.GET("/:id/download", resumeH.GetDownloadLink) // GET /api/resumes/:id/download
		resumes.PUT("/:id/status", resumeH.UpdateStatus)      // PUT /api/resumes/:id/status
		resumes.DELETE("/:id", resumeH.Delete)                // DELETE /api/resumes/:id
	}

	// Interview admin routes
	interviews := api.Group("/interviews")
	{
		interviews.POST("", interviewH.Create) // POST /api/interviews
	}

	// Public interview routes (для кандидатов)
	public := r.Group("/interview")
	{
		public.GET("/:token", interviewH.GetByToken)            // GET /interview/:token
		public.POST("/:token/start", interviewH.StartByToken)   // POST /interview/:token/start
		public.POST("/:token/finish", interviewH.FinishByToken) // POST /interview/:token/finish
	}

	return r
}
