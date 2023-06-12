package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

type Message struct {
	Message   string
	Sender    string
	Timestamp int64
}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	err := ValidateSendRequest(req)
	if err != nil {
		return nil, err
	} else {
		fmt.Println("Send Request is valid, proceed to create message")
	}

	message := createMessage(req)
	fmt.Println("Message created as:", message)

	// TODO: generate key
	groupID, err := checkGroupID(req.Message.GetChat())
	if err != nil {
		return nil, err
	}
	fmt.Println("Group ID is:", groupID)

	// TODO: save message to database
	err = rdb.SaveMessageToRedis(ctx, groupID, message)
	if err != nil {
		return nil, err
	}
	fmt.Println("Message saved to redis")

	resp := rpc.NewSendResponse()
	resp.Code = 0
	resp.Msg = fmt.Sprintf("Success: Sent message %s", req.Message.GetChat())
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {

	fmt.Println("Pull Request received:", req)
	if req.GetChat() == "" {
		return nil, fmt.Errorf("chat ID is required")
	}

	groupID, err := checkGroupID(req.GetChat())
	if err != nil {
		return nil, err
	}
	fmt.Println("Group ID is:", groupID)

	limit := int64(req.GetLimit())
	if limit == 0 {
		limit = 10 // default limit 10
	}

	start := req.GetCursor()
	end := start + limit - 1

	messages, err := rdb.GetMessagesByGroupID(ctx, groupID, start, end, req.GetReverse())
	if err != nil {
		return nil, err
	}
	fmt.Println("Messages retrieved from redis:", messages)

	respMessages := make([]*rpc.Message, 0)
	hasMore := false
	if len(messages) > int(limit) {
		messages = messages[:limit]
		hasMore = true
	}

	for _, msg := range messages {
		temp := &rpc.Message{
			Chat:     req.Chat, // Use req.Chat instead of req.GetChat()
			Text:     msg.Message,
			Sender:   msg.Sender,
			SendTime: msg.Timestamp,
		}
		respMessages = append(respMessages, temp)
	}

	nextCursor := start + limit
	resp := rpc.NewPullResponse()
	resp.Messages = respMessages
	resp.Code = 0
	resp.Msg = "success"
	resp.HasMore = &hasMore
	resp.NextCursor = &nextCursor

	return resp, nil
}

func ValidateSendRequest(req *rpc.SendRequest) error {
	chatID := strings.Split(req.Message.GetChat(), ":")
	message := req.Message.GetText()

	// validate chat id
	if len(chatID) == 0 {
		return fmt.Errorf("'%s' is an invalid chat id, chat id should not be empty", req.Message.GetChat())
	} else if len(chatID) < 2 {
		return fmt.Errorf("'%s' is an invalid chat id, chat id should be in the format of sender:receiver", req.Message.GetChat())
	}

	// validate message
	if len(message) == 0 {
		return fmt.Errorf("message should not be empty")
	}

	// validate sender
	sender1, sender2 := chatID[0], chatID[1]
	if req.Message.GetSender() != sender1 && req.Message.GetSender() != sender2 {
		return fmt.Errorf("'%s' is an invalid sender, sender is not in chat", req.Message.GetSender())
	}
	return nil
}

func min(a, b string) string {
	if strings.Compare(a, b) <= 0 {
		return a
	}
	return b
}

func max(a, b string) string {
	if strings.Compare(a, b) >= 0 {
		return a
	}
	return b
}

func checkGroupID(groupID string) (string, error) {
	lowercase := strings.ToLower(groupID)
	senders := strings.Split(lowercase, ":")
	if len(senders) != 2 {
		return "", fmt.Errorf("invalid Group ID '%s', should be in the format of user1:user2 test", groupID)
	}

	sender1, sender2 := senders[0], senders[1]
	// Compare the sender and receiver alphabetically and sort them to form the Group ID
	groupID = fmt.Sprintf("%s:%s", min(sender1, sender2), max(sender1, sender2))

	return groupID, nil
}

func createMessage(req *rpc.SendRequest) *Message {
	return &Message{
		Message:   req.Message.GetText(),
		Sender:    req.Message.GetSender(),
		Timestamp: time.Now().Unix(),
	}
}
