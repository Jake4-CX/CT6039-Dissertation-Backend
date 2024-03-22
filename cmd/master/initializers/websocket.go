package initializers

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
	"github.com/zishang520/engine.io/v2/utils"
	"github.com/zishang520/socket.io/v2/socket"
)

var SocketIO *socket.Server

func InitializeWebsocket() {
	SocketIO = socket.NewServer(nil, nil).Of(regexp.MustCompile(`/\w+`), nil).Use(func(client *socket.Socket, next func(*socket.ExtendedError)) {
		utils.Log().Success("MId:%v", client.Connected())
		next(nil)
	}).Server()

	SocketIO.On("connection", onConnect)

	log.Info("Socket.IO server initialized")

}

func onConnect(clients ...interface{}) {
	client := clients[0].(*socket.Socket)
	log.Info("Connected: ", client.Id())
	client.Emit("connection", "Successfully connected to WebSocket. :)")

	client.Join("test")

	client.On("subscribeTest", func(args ...interface{}) {
		subscribeTest(client, args)
	})

	client.On("unsubscribeTest", func(args ...interface{}) {
		unsubscribeTest(client, args)
	})

	client.On("disconnect", func(reason ...any) {
		onDisconnect(client, reason)
	})
}

func onDisconnect(client *socket.Socket, reason ...any) {
	log.Infof("Disconnected: %v, Reason: %v", client.Id(), reason)
}

func subscribeTest(client *socket.Socket, args ...interface{}) {
	var subscribeTest structs.SubscribeTest
	if err := decodeArgsToStruct(args, &subscribeTest); err != nil {
		log.Errorf("Error decoding to SubscribeTest: %v", err)
		client.Emit("error", "Internal server error")
		return
	}

	log.Infof("Successfully decoded: %+v", subscribeTest)

	roomName := socket.Room("loadTest:" + strconv.Itoa(subscribeTest.TestId))
	client.Join(roomName)
	client.Emit("joinedRoom", roomName)
}

func unsubscribeTest(client *socket.Socket, args ...interface{}) {
	var unsubscribeTest structs.UnsubscribeTest
	if err := decodeArgsToStruct(args, &unsubscribeTest); err != nil {
		log.Errorf("Error decoding to UnsubscribeTest: %v", err)
		client.Emit("error", "Internal server error")
		return
	}

	log.Infof("Successfully decoded: %+v", unsubscribeTest)

	roomName := socket.Room("loadTest:" + strconv.Itoa(unsubscribeTest.TestId))

	if !client.Rooms().Has(roomName) {
		client.Emit("error", "You are not subscribed to this test")
		return
	}

	client.Leave(roomName)

	client.Emit("unsubscribed", map[string]interface{}{"testId": unsubscribeTest.TestId})
}

func decodeArgsToStruct(args []interface{}, result interface{}) error {
	if len(args) != 1 {
		return errors.New("invalid arguments")
	}

	innerSlice, ok := args[0].([]interface{})
	if !ok || len(innerSlice) != 1 {
		return errors.New("invalid arguments format")
	}

	dataMap, ok := innerSlice[0].(map[string]interface{})
	if !ok {
		return errors.New("expected a map in arguments")
	}

	err := mapstructure.Decode(dataMap, result)
	if err != nil {
		return err
	}

	return nil
}
