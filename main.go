package main

import (
	"alice_gorpc/msg"
	"alice_gorpc/protocol"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":4999")
	if err != nil {
		panic(err)
	}
	go ClientWrite()
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		frame := protocol.NewFrame()
		data, err := frame.ReadFrame(conn)
		if err != nil {
			panic(err)
		}
		codec := protocol.NewCodec()
		body := codec.Decode(data)
		req := msg.Request{}
		_ = json.Unmarshal(body, req)
		rsp := &msg.Response{Payload: []byte("hhhhh")}
		response, err := json.Marshal(rsp)
		if err != nil {
			panic(err)
		}
		response, err = codec.Encode(response)
		if err != nil {
			panic(err)
		}
		_, err = conn.Write(response)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}

func ClientWrite() {
	conn, err := net.Dial("tcp", "127.0.0.1:4999")
	if err != nil {
		panic(err)
	}
	for {
		fmt.Println("---------")
		req := &msg.Request{
			ServicePath: "",
			Metadata:    nil,
			Payload:     []byte("hello"),
		}
		body, _ := json.Marshal(req)
		codec := protocol.NewCodec()
		frame, err := codec.Encode(body)
		if err != nil {
			panic(err)
		}
		_, err = conn.Write(frame)
		if err != nil {
			panic(err)
		}
		rspFrame := make([]byte, 1000)

		_, err = conn.Read(rspFrame)
		if err != nil {
			fmt.Println("conn read err:", err)
			panic(err)
		}
		fmt.Println(rspFrame)
		time.Sleep(1 * time.Second)
	}
}
