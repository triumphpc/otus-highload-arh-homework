package dialog

import (
	"context"
	"errors"
	"log"
	"time"

	dialogv1 "otus-highload-arh-homework/pkg/proto/dialog/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn    *grpc.ClientConn
	client  dialogv1.DialogServiceClient
	timeout time.Duration
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(requestIDInterceptor),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err = errors.Join(err, conn.Close())
	}(conn)

	return &Client{
		conn:    conn,
		client:  dialogv1.NewDialogServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SendMessage(ctx context.Context, senderID, receiverID, text string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.SendMessage(ctx, &dialogv1.SendMessageRequest{
		SenderId:   senderID,
		ReceiverId: receiverID,
		Text:       text,
	})
	return err
}

func (c *Client) GetMessages(ctx context.Context, userID, otherUserID string) ([]*dialogv1.DialogMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetMessages(ctx, &dialogv1.GetMessagesRequest{
		UserId:      userID,
		OtherUserId: otherUserID,
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}

func requestIDInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// Получаем request-id из контекста Gin
	if requestID, ok := ctx.Value("x-request-id").(string); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
