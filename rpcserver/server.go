// Package rpcserver provides a net/rpc server and client for controlling a
// running Adele application out of band, such as toggling maintenance mode.
//
// The server listens on a configurable TCP address and can be disabled via the
// RPC_SERVER_DISABLE environment variable.
package rpcserver

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/cidekar/adele-framework"
)

const (
	ServerAddrDefault = "127.0.0.1"
	ServerPortDefault = "4040"
)

// Server
type RPCServer struct {
	App *adele.Adele
}

func (r *RPCServer) SetMaintenanceMode(args *MaintenanceModeArgs, reply *MaintenanceModeReply) error {
	r.App.MaintenanceMode = args.InMaintenanceMode
	if r.App.MaintenanceMode {
		reply.Status = "down"
	} else {
		reply.Status = "up"
	}
	return nil
}

func Start(app *adele.Adele) error {

	if adele.Helpers.Getenv("RPC_SERVER_DISABLE") != "" {
		return nil
	}

	rs := &RPCServer{
		App: app,
	}

	err := rpc.Register(rs)
	if err != nil {
		return fmt.Errorf("failed to publish the reciever: %s", err)
	}

	ServerAddr := adele.Helpers.Getenv("RPC_SERVER_ADDR", ServerAddrDefault)
	ServerPort := adele.Helpers.Getenv("RPC_SERVER_PORT", ServerPortDefault)

	listener, err := net.Listen("tcp", ServerAddr+":"+ServerPort)
	if err != nil {
		return fmt.Errorf("failed to announce on the local network address: %s", err)
	}

	app.RPCListener = &listener

	// Start accepting connections in a goroutine so this function can return
	go func() {
		defer listener.Close() // Ensure cleanup if goroutine exits
		for {
			rpcConn, err := listener.Accept()
			if err != nil {
				// Check if listener was closed intentionally
				if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
					return // Exit gracefully
				}
				continue
			}
			go rpc.ServeConn(rpcConn)
		}
	}()

	return nil

}

func Stop(app *adele.Adele) error {
	if adele.Helpers.Getenv("RPC_SERVER_DISABLE") != "" {
		return nil
	}

	if app == nil {
		return fmt.Errorf("can not close rpc listener on a nil application")
	}

	if app.RPCListener == nil {
		return nil
	}

	err := (*app.RPCListener).Close()
	if err != nil {
		return fmt.Errorf("failed to close RPC listener: %w", err)
	}

	(*app.RPCListener) = nil

	return nil

}
