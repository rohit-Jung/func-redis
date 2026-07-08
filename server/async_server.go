package server

import (
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rohit-Jung/func-redis/config"
	"github.com/rohit-Jung/func-redis/core"
)

var connClients int = 0
var maxClients int = 20_000

var deletionFrequency time.Duration = 1 * time.Second
var lastDeletedTime time.Time = time.Now()

const EngineStatus_WAITING int32 = 1 << 1
const EngineStatus_BUSY int32 = 1 << 2
const EngineStatus_SHUTTING_DOWN int32 = 1 << 2

var eStatus int32 = EngineStatus_WAITING

func SignalHandelling(wg *sync.WaitGroup, sigs chan os.Signal) {
	// defer done and keep receiving signals
	defer wg.Done()
	<-sigs

	// wait if it is busy
	// why do you need load ?
	for atomic.LoadInt32(&eStatus) == EngineStatus_BUSY {
	}

	// CRITICAL TO HANDLE
	// the server should not go to BUSY state between these two

	// set the status to SHUTTING DOWN (only place we can do that)
	atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)

	// shutdown
	core.Shutdown()
	os.Exit(0)
}

func RunAsyncServer(wg *sync.WaitGroup) error {
	// when the function is done, set the status to shutting down
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)
	}()

	log.Println("starting async tcp server on", config.Host, config.Port)

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

	// we want to loop unless the Engine is SHUTTING_DOWN
	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTTING_DOWN {

		// last deleted time has passed
		if time.Now().After(lastDeletedTime.Add(deletionFrequency)) {
			core.DeleteExpiredKeys()
			lastDeletedTime = time.Now()
		}

		// what is -1 here ?
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		// do  not break the loop
		if err != nil {
			continue
		}

		// after Epoll Wait file descripter is ready with an io so
		// we want to set the engine state to busy here

		// but what if the state is already shutting down
		// hence we use compare and swap, only if the state is waiting we make it busy
		swapSuccessful := atomic.CompareAndSwapInt32(&eStatus, EngineStatus_WAITING, EngineStatus_BUSY)

		// if we could not swap meaning the state was either already busy or shutting down
		if !swapSuccessful {
			// so change it to shutting down
			switch eStatus {
			// if it was already shutting down then do not go below on to processing the IO events
			case EngineStatus_SHUTTING_DOWN:
				return nil
			}

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
				cmds, err := core.ReadCommands(comm)
				if err != nil {
					log.Print("Error reading command")
					connClients -= 1
					comm.Close()
					continue
				}

				cmds.Respond(comm)
			}
		}

		// processing is done so now we can mark the engine status to be waiting again
		atomic.StoreInt32(&eStatus, EngineStatus_WAITING)
	}

	return nil
}
