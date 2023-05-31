package model

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

type BanTx struct {
	UserName string `json:"username"`
}

// Message represents a message sent by a user
type Message struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

type MsgHistory struct {
	Msg string `json:"history"`
}

func AppendHistory(db *DB, message Message) error {
	historyBytes, err := ViewDB(db.GetDB(), []byte("history"))
	if err != nil {
		fmt.Println("Error fething historu:", err)
		return err
	}
	msgBytes := string(historyBytes)
	msgBytes = msgBytes + "{sender:" + message.Sender + ",message:" + message.Message + "}"
	if err != nil {
		fmt.Println("erro appending mssg to history: ", err)
	}

	fmt.Println("Appending message to history")
	err = db.Set([]byte("history"), []byte(msgBytes))
	if err != nil {
		fmt.Println("Error updating history:", err)
	}
	return err
}

func FetchHistory(db *DB) (string, error) {
	historyBytes, err := ViewDB(db.GetDB(), []byte("history"))
	if err != nil {
		fmt.Println("Error fething historu:", err)
		return "", err
	}
	msgHistory := string(historyBytes)

	if err != nil {
		fmt.Println("erro appending history: ", err)
	}
	return msgHistory, err
}

// AddMessage adds a message to the database
// AddMessage uses string
// To be changed to array later
func AddMessage(db *DB, message Message) error {
	// Get the existing messages for the sender
	existingMessages, err := GetMessagesBySender(db, message.Sender)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	// Append the new message to the string
	fmt.Println("existing Pre :", existingMessages)
	existingMessages = existingMessages + ";" + message.Message
	fmt.Println("existing Post :", existingMessages)
	// Store the updated string in the Badger database
	err = db.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(message.Sender+"msg"), []byte(existingMessages))
	})
	if err != nil {
		return err
	}
	err = AppendHistory(db, message)
	return err
}

// GetMessagesBySender retrieves all messages sent by a specific sender
// Get Message using String
func GetMessagesBySender(db *DB, sender string) (string, error) {
	var messages string
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(sender + "msg"))
		if err != nil {
			return err
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		messages = string(value)
		return nil
	})
	if err != nil {
		return "", err
	}
	return messages, nil
}

// Parse Message
func ParseMessage(tx []byte) (*Message, error) {
	msg := &Message{}

	// Parse the message into key-value pairs
	pairs := strings.Split(string(tx), ",")

	if len(pairs) != 2 {
		return nil, errors.New("invalid number of key-value pairs in message")
	}

	for _, pair := range pairs {
		kv := strings.Split(pair, ":")

		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key-value pair in message: %s", pair)
		}

		key := kv[0]
		value := kv[1]

		switch key {
		case "sender":
			msg.Sender = value
		case "message":
			msg.Message = value
		default:
			return nil, fmt.Errorf("unknown key in message: %s", key)
		}
	}

	// Check if the message contains a sender and message
	if msg.Sender == "" {
		return nil, errors.New("message is missing sender")
	}
	if msg.Message == "" {
		return nil, errors.New("message is missing message")
	}

	return msg, nil
}
