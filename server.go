package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms map[string]*room
	commands chan command

}

func newServer() *server {
	return &server{
		rooms: make(map[string]*room),
		commands: make(chan command),
	}
}

func(s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NICK:
			s.nick(cmd.client, cmd.args[1])
		case CMD_JOIN:
			s.join(cmd.client, cmd.args[1])
		case CMD_ROOMS:
			s.listRooms(cmd.client)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client)
		}
	}
}
func (s *server) newClient(conn net.Conn) {
	log.Printf("New Client has Connected: %s", conn.RemoteAddr().String())

	c := &client{
		conn: 		conn,
		nick: 		"anonymous",
		commands: 	s.commands,
	}

	c.readInput()
}

func (s *server) nick(c *client, nick string){
	c.nick = nick
	c.msg(fmt.Sprintf("Your name has been set to %s", nick)) 
}
func (s *server) join(c *client, roomName string){
	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name: roomName,
			members: make(map[net.Addr]*client),
		}
		s.rooms[roomName] = r
	}

	r.members[c.conn.RemoteAddr()] = c
	
	s.quitCurrentRoom(c)

	c.room = r

	r.broadcast(c, fmt.Sprintf("%s Joined the room", c.nick))
	c.msg(fmt.Sprintf("Welcome to %s", roomName))
}
func (s *server) listRooms(c *client){
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	c.msg(fmt.Sprintf("Available rooms are: %s", strings.Join(rooms, ", ")))
}
func (s *server) msg(c *client, args []string){
	msg := strings.Join(args[1:len(args)], " ")
	c.room.broadcast(c, c.nick +": " + msg)
}
func (s *server) quit(c *client){
	log.Printf("Client disconnected: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)

	c.msg("Goodbye, disconnecting")
	c.conn.Close()
}

func (s *server) quitCurrentRoom(c *client){
	if c.room != nil{
		oldRoom := s.rooms[c.room.name]
		delete(s.rooms[c.room.name].members, c.conn.RemoteAddr())
		oldRoom.broadcast(c, fmt.Sprintf("%s has left the room", c.nick))

	}
}
