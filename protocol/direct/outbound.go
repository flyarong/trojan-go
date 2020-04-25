package direct

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type DirectOutboundConnSession struct {
	protocol.ConnSession
	conn    io.ReadWriteCloser
	request *protocol.Request
}

func (o *DirectOutboundConnSession) Read(p []byte) (int, error) {
	return o.conn.Read(p)
}

func (o *DirectOutboundConnSession) Write(p []byte) (int, error) {
	return o.conn.Write(p)
}

func (o *DirectOutboundConnSession) Close() error {
	return o.conn.Close()
}

func NewOutboundConnSession(req *protocol.Request) (protocol.ConnSession, error) {
	newConn, err := net.Dial(req.Network(), req.String())
	if err != nil {
		return nil, err
	}
	o := &DirectOutboundConnSession{
		request: req,
		conn:    newConn,
	}
	return o, nil
}

type packetInfo struct {
	request *protocol.Request
	packet  []byte
}

type DirectOutboundPacketSession struct {
	protocol.PacketSession
	packetChan chan *packetInfo
	ctx        context.Context
	cancel     context.CancelFunc
}

func (o *DirectOutboundPacketSession) listenConn(req *protocol.Request, conn *net.UDPConn) {
	defer conn.Close()
	for {
		buf := make([]byte, protocol.MaxUDPPacketSize)
		conn.SetReadDeadline(time.Now().Add(protocol.UDPTimeout))
		n, addr, err := conn.ReadFromUDP(buf)
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			log.Info(err)
			return
		}
		if addr.String() != req.String() {
			panic("addr != req, something went wrong")
		}
		info := &packetInfo{
			request: req,
			packet:  buf[0:n],
		}
		o.packetChan <- info
	}
}

func (o *DirectOutboundPacketSession) Close() error {
	o.cancel()
	return nil
}

func (o *DirectOutboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	select {
	case info := <-o.packetChan:
		return info.request, info.packet, nil
	case <-o.ctx.Done():
		return nil, nil, common.NewError("session closed")
	}
}

func (o *DirectOutboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	var remote *net.UDPAddr
	if req.AddressType == common.DomainName {
		remote, err := net.ResolveUDPAddr("", string(req.DomainName))
		if err != nil {
			return 0, err
		}
		remote.Port = req.Port
	} else {
		remote = &net.UDPAddr{
			IP:   req.IP,
			Port: req.Port,
		}
	}
	conn, err := net.DialUDP("udp", nil, remote)
	if err != nil {
		return 0, common.NewError("cannot dial udp").Base(err)
	}
	log.Debug("udp directly dialing to", remote)
	go o.listenConn(req, conn)
	n, err := conn.Write(packet)
	return n, err
}

func NewOutboundPacketSession(ctx context.Context) (protocol.PacketSession, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &DirectOutboundPacketSession{
		ctx:        ctx,
		cancel:     cancel,
		packetChan: make(chan *packetInfo, 256),
	}, nil
}
