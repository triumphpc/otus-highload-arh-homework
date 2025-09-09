package prometheus

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// RED метрики
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "path", "status"},
	)

	// бизнес метрики
	userRegistrationsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_user_registrations_total",
			Help: "Total number of user registrations",
		},
	)
	userLoginAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_user_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"success"}, // "true" или "false"
	)
	postsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_posts_created_total",
			Help: "Total number of posts created",
		},
	)
	messagesSentTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_messages_sent_total",
			Help: "Total number of dialog messages sent",
		},
		[]string{"from_user_id", "to_user_id"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(userRegistrationsTotal)
	prometheus.MustRegister(userLoginAttemptsTotal)
	prometheus.MustRegister(postsCreatedTotal)
	prometheus.MustRegister(messagesSentTotal)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method

		httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(method, path, strconv.Itoa(status)).Observe(duration)
	}
}

// PUBLIC METHODS FOR BUSINESS METRICS

// IncUserRegistrations увеличивает счетчик регистраций пользователей
func IncUserRegistrations() {
	userRegistrationsTotal.Inc()
}

// IncLoginAttempts увеличивает счетчик попыток входа
// success: true - успешный вход, false - неуспешный
func IncLoginAttempts(success bool) {
	status := "false"
	if success {
		status = "true"
	}
	userLoginAttemptsTotal.WithLabelValues(status).Inc()
}

// IncPostsCreated увеличивает счетчик созданных постов
func IncPostsCreated() {
	postsCreatedTotal.Inc()
}

// IncMessagesSent увеличивает счетчик отправленных сообщений
func IncMessagesSent(fromUserID, toUserID string) {
	messagesSentTotal.WithLabelValues(fromUserID, toUserID).Inc()
}
