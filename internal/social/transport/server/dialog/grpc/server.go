package grpc

import (
	"context"
	"net"

	userUC "otus-highload-arh-homework/internal/social/usecase/user"
	"otus-highload-arh-homework/pkg/proto/dialog/v1"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server       *grpc.Server
	lis          net.Listener
	healthServer *health.Server
}

func New(uc *userUC.UserUseCase, port string) (*Server, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			requestIDInterceptor,
			loggingInterceptor,
		),
	)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, healthServer)

	// Регистрируем сервис
	dialogv1.RegisterDialogServiceServer(srv, NewDialogService(uc))

	// Для разработки - reflection API
	reflection.Register(srv)

	return &Server{
		server:       srv,
		lis:          lis,
		healthServer: healthServer,
	}, nil
}

func (s *Server) Run() error {
	return s.server.Serve(s.lis)
}

func (s *Server) Stop() {
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.server.GracefulStop()
}

func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	// Получаем request-id из контекста
	requestID, ok := ctx.Value("x-request-id").(string)
	fields := logrus.Fields{
		"method": info.FullMethod,
	}
	if ok {
		fields["x-request-id"] = requestID
	}

	// Логируем входящий запрос
	logrus.WithFields(fields).Info("gRPC method called")

	// Продолжаем выполнение
	return handler(ctx, req)
}

func requestIDInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		requestIDs := md.Get("x-request-id")
		if len(requestIDs) > 0 {
			requestID := requestIDs[0]

			logrus.WithFields(logrus.Fields{
				"x-request-id": requestID,
				"method":       info.FullMethod,
			}).Info("gRPC request")

			// Добавляем request-id в контекст для дальнейшего использования
			ctx = context.WithValue(ctx, "x-request-id", requestID)
		}
	}

	// Продолжаем выполнение
	return handler(ctx, req)
}
