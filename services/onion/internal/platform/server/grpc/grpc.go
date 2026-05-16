package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/voxire/onion/internal/platform/logger"
)

type Server struct {
	server *grpc.Server
	port   string
}

func NewServer(port string) *Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
		grpc.StreamInterceptor(streamLoggingInterceptor),
	)

	reflection.Register(server)

	return &Server{
		server: server,
		port:   port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lc := net.ListenConfig{}

	lis, err := lc.Listen(ctx, "tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.Ctx(ctx).Info("Starting gRPC server", logger.String("port", s.port))

	go func() {
		if err := s.server.Serve(lis); err != nil {
			logger.Ctx(ctx).Error("gRPC server failed", logger.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Ctx(ctx).Info("Shutting down gRPC server...")
		s.server.GracefulStop()
	}()

	return nil
}

func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger.Ctx(ctx).Info("gRPC request",
		logger.String("method", info.FullMethod),
		logger.Any("request", req),
	)

	resp, err := handler(ctx, req)
	if err != nil {
		logger.Ctx(ctx).Error("gRPC error",
			logger.String("method", info.FullMethod),
			logger.Error(err),
		)
	} else {
		logger.Ctx(ctx).Info("gRPC response",
			logger.String("method", info.FullMethod),
		)
	}

	return resp, err
}

func streamLoggingInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logger.Ctx(ss.Context()).Info("gRPC stream request",
		logger.String("method", info.FullMethod),
	)

	err := handler(srv, ss)
	if err != nil {
		logger.Ctx(ss.Context()).Error("gRPC stream error",
			logger.String("method", info.FullMethod),
			logger.Error(err),
		)
	}

	return err
}
