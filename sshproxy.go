package sshproxy

import (
	"code.google.com/p/go.crypto/ssh"
	"io"
	"log"
	"net"
)

type SshConn struct {
	net.Conn
	config     *ssh.ServerConfig
	callbackFn func(c ssh.ConnMetadata) (*ssh.Client, error)
	wrapFn     func(c ssh.ConnMetadata, r io.ReadCloser) (io.ReadCloser, error)
	closeFn    func(c ssh.ConnMetadata) error
}

func (p *SshConn) serve() error {
	serverConn, chans, reqs, err := ssh.NewServerConn(p, p.config)
	if err != nil {
		log.Println("failed to handshake")
		return (err)
	}

	defer serverConn.Close()

	clientConn, err := p.callbackFn(serverConn)
	if err != nil {
		log.Printf("%s", err.Error())
		return (err)
	}

	defer clientConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {

		channel2, requests2, err2 := clientConn.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
		if err2 != nil {
			log.Fatalf("Could not accept client channel: %s", err.Error())
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept server channel: %s", err.Error())
		}

		// connect requests
		go func() {
			log.Printf("Waiting for request")

		r:
			for {
				var req *ssh.Request
				var dst ssh.Channel

				select {
				case req = <-requests:
					dst = channel2
				case req = <-requests2:
					dst = channel
				}

				log.Printf("Request: %s %s %s %s\n", dst, req.Type, req.WantReply, req.Payload)

				b, err := dst.SendRequest(req.Type, req.WantReply, req.Payload)
				if err != nil {
					log.Printf("%s", err)
				}

				if req.WantReply {
					req.Reply(b, nil)
				}

				switch req.Type {
				case "exit-status":
					break r
				case "exec":
					// not supported (yet)
				default:
					log.Println(req.Type)
				}
			}

			channel.Close()
			channel2.Close()
		}()

		// connect channels
		log.Printf("Connecting channels.")

		var wrappedChannel io.ReadCloser = channel
		var wrappedChannel2 io.ReadCloser = channel2

		if p.wrapFn != nil {
			// wrappedChannel, err = p.wrapFn(channel)
			wrappedChannel2, err = p.wrapFn(serverConn, channel2)
		}

		go io.Copy(channel2, wrappedChannel)
		go io.Copy(channel, wrappedChannel2)

		defer wrappedChannel.Close()
		defer wrappedChannel2.Close()
	}

	if p.closeFn != nil {
		p.closeFn(serverConn)
	}

	return nil
}

func ListenAndServe(addr string, serverConfig *ssh.ServerConfig,
	callbackFn func(c ssh.ConnMetadata) (*ssh.Client, error),
	wrapFn func(c ssh.ConnMetadata, r io.ReadCloser) (io.ReadCloser, error),
	closeFn func(c ssh.ConnMetadata) error,
) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("listen.Accept failed: %v", err)
		}

		sshconn := &SshConn{Conn: conn, config: serverConfig, callbackFn: callbackFn, wrapFn: wrapFn, closeFn: closeFn}

		go func() {
			if err := sshconn.serve(); err != nil {
				log.Printf("Error occured while serving %s\n", err)
				return
			}

			log.Println("Connection closed.")
		}()
	}

}
