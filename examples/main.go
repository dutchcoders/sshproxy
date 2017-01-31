package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/Unknwon/com"
	"github.com/dutchcoders/sshproxy"
	"golang.org/x/crypto/ssh"
)

func main() {
	listen := flag.String("listen", ":8022", "listen address")
	dest := flag.String("dest", ":22", "destination address")
	key := flag.String("key", "conf/id_rsa", "rsa key to use")
	flag.Parse()

	path := *key
	privateBytes, err := ssh_keyGen(path)
	if err != nil {
		panic(err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	var sessions map[net.Addr]map[string]interface{} = make(map[net.Addr]map[string]interface{})

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			fmt.Printf("Login attempt: %s, user %s password: %s", c.RemoteAddr(), c.User(), string(pass))

			sessions[c.RemoteAddr()] = map[string]interface{}{
				"username": c.User(),
				"password": string(pass),
			}

			clientConfig := &ssh.ClientConfig{}

			clientConfig.User = c.User()
			clientConfig.Auth = []ssh.AuthMethod{
				ssh.Password(string(pass)),
			}

			client, err := ssh.Dial("tcp", *dest, clientConfig)

			sessions[c.RemoteAddr()]["client"] = client
			return nil, err
		},
	}

	config.AddHostKey(private)

	sshproxy.ListenAndServe(*listen, config, func(c ssh.ConnMetadata) (*ssh.Client, error) {
		meta, _ := sessions[c.RemoteAddr()]

		fmt.Println(meta)

		client := meta["client"].(*ssh.Client)
		fmt.Printf("Connection accepted from: %s", c.RemoteAddr())

		return client, err
	}, func(c ssh.ConnMetadata, r io.ReadCloser) (io.ReadCloser, error) {
		return sshproxy.NewTypeWriterReadCloser(r), nil
	}, func(c ssh.ConnMetadata) error {
		fmt.Println("Connection closed.")
		return nil
	})
}

func ssh_keyGen(keyPath string) ([]byte, error) {
	var privateBytes []byte
	var err error
	if !com.IsExist(keyPath) {
		priv, err := rsa.GenerateKey(rand.Reader, 2014)
		if err != nil {
			fmt.Println(err)
			return privateBytes, err
		}
		err = priv.Validate()
		if err != nil {
			fmt.Println("Validation failed.", err)
			return privateBytes, err
		}

		priv_der := x509.MarshalPKCS1PrivateKey(priv)

		priv_blk := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   priv_der,
		}

		privateBytes = pem.EncodeToMemory(&priv_blk)
		return privateBytes, err
	}

	if privateBytes, err = ioutil.ReadFile(keyPath); err != nil {
		panic("Failed to load private key")
	}
	return privateBytes, err
}
