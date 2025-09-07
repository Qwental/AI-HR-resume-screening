package handlers

import (
	"github.com/gin-gonic/gin"
	"interview/internal/service"
)

func SetupRouter(
	vacancySvc service.VacancyService,
	resumeSvc service.ResumeService,
	interviewSvc service.InterviewService,
	aiSvc service.AIService,
	chatSvc service.ChatService,
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

	// Создаём handlers
	vacancyH := NewVacancyHandler(vacancySvc)
	resumeH := NewResumeHandler(resumeSvc)
	interviewH := NewHandler(interviewSvc)
	chatH := NewChatHandler(chatSvc, interviewSvc)

	api := r.Group("/api")
	{
		// Vacancy routes
		vacancies := api.Group("/vacancies")
		{
			vacancies.GET("", vacancyH.GetAll)              // GET /api/vacancies (все вакансии)
			vacancies.POST("", vacancyH.Create)             // POST /api/vacancies
			vacancies.GET("/:id", vacancyH.GetDownloadLink) // GET /api/vacancies/:id/download
			vacancies.PUT("/:id", vacancyH.Update)          // PUT /api/vacancies/:id
			vacancies.DELETE("/:id", vacancyH.Delete)       // DELETE /api/vacancies/:id

			vacancies.GET("/resume/:vacancy_id", resumeH.GetByVacancy) // GET /api/vacancies/:vacancy_id/resumes
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

		// Admin interview routes
		admin := api.Group("/admin/interviews")
		{
			admin.POST("", interviewH.Create) // POST /api/admin/interviews
		}
	}

	// Public routes (для кандидатов)
	public := r.Group("/interview")
	{
		// Существующие роуты интервью
		public.GET("/:token", interviewH.GetByToken)            // GET /interview/:token
		public.POST("/:token/start", interviewH.StartByToken)   // POST /interview/:token/start
		public.POST("/:token/finish", interviewH.FinishByToken) // POST /interview/:token/finish

		// Новые чат-роуты
		public.POST("/:token/message", chatH.SendMessage) // POST /interview/:token/message
		public.GET("/:token/messages", chatH.GetMessages) // GET /interview/:token/messages
		public.GET("/:token/status", chatH.GetStatus)     // GET /interview/:token/status
	}

	return r
}
