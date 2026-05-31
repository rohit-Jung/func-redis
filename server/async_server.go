package server

import (
	"log"
	"net"
	"syscall"

	"github.com/rohit-Jung/func-redis/config"
	"github.com/rohit-Jung/func-redis/core"
)

func RunAsyncServer() error {
	log.Println("starting async tcp server on", config.Host, config.Port)

	maxClients := 20_000
	connClients := 0

	// what is the last arg proto ?
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}

	defer syscall.Close(serverFD)

	// what is the true here
	if err := syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// bind and listen
	ipv4 := net.ParseIP(config.Host)
	sockAddr := &syscall.SockaddrInet4{
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
		Port: config.Port,
	}

	if err := syscall.Bind(serverFD, sockAddr); err != nil {
		return err
	}

	// backlog of max_clients
	if err := syscall.Listen(serverFD, maxClients); err != nil {
		return err
	}

	// async io
	// create an epoll fd

	// flag 0 - similar to epoll_create
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epollFD)

	// create server event to monitor
	serverEvent := &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// register the server fd in the epollFD
	if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, serverEvent); err != nil {
		return err
	}

	// make events to hold new incoming events
	events := make([]syscall.EpollEvent, maxClients)
	for {
		// what is -1 here ?
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		// do  not break the loop
		if err != nil {
			continue
		}

		for i := range nevents {
			// meaning its a tcp connection so accept it
			if int(events[i].Fd) == serverFD {
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					continue
				}

				connClients++
				// set this fd also to non blocking
				syscall.SetNonblock(fd, true)
				// monitor this fd for events too
				socketEvent := &syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, socketEvent); err != nil {
					log.Print("Error adding new fd to epoll fd")
					continue
				}
			} else {
				// else client is request for io
				comm := &core.FDComm{
					Fd: int(events[i].Fd),
				}

				// read command and respond
				cmd, err := core.ReadCommand(comm)
				if err != nil {
					log.Print("Error reading command")
					connClients -= 1
					comm.Close()
					continue
				}

				cmd.Respond(comm)
			}

		}

	}

}
