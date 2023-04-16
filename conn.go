package rtm2_sdk

import (
	"context"
	"encoding/binary"
	"errors"
	"github.com/cloudwego/netpoll"
	"github.com/tevino/abool/v2"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

const defaultChannelSize = 4096
const defaultFlushSize = 1024 * 1024
const serviceSubproxy = 2380 // todo change service id

type Marshalable interface {
	MarshalTo([]byte) (int, error)
	Marshal() ([]byte, error)
	Size() int
}

type connectionCallback interface {
	onResponse(uri int32, errCode int32, message []byte) error
}

type connection struct {
	ctx      context.Context
	cancel   context.CancelFunc
	lg       *zap.Logger
	callback connectionCallback
	edp      string
	req      chan *Header
	resp     chan *Header
	requests sync.Map
	seqId    int64
	errChan  chan error
	start    abool.AtomicBool
	conn     netpoll.Connection
}

func NewConnection(gCtx context.Context, lg *zap.Logger, edp string, callback connectionCallback) *connection {
	ctx, cancel := context.WithCancel(gCtx)
	ret := &connection{
		ctx:      ctx,
		cancel:   cancel,
		lg:       lg.With(zap.String("endpoint", edp)),
		callback: callback,
		edp:      edp,
		req:      make(chan *Header, defaultChannelSize),
		resp:     make(chan *Header, defaultChannelSize),
		errChan:  make(chan error, 10),
	}

	return ret
}

func (c *connection) Start() {
	if c.start.SetToIf(false, true) {
		go c.loop()
	}
}

func (c *connection) SendRequest(header *Header, rc chan<- *Header) error {
	seqId := atomic.AddInt64(&c.seqId, 1)
	if rc != nil {
		c.requests.Store(seqId, rc)
	}
	header.SeqId = seqId
	select {
	case c.req <- header:
		return nil
	default:
		c.lg.Warn("channel full")
		return errors.New("channel full")
	}
}

func (c *connection) ErrorChan() <-chan error {
	return c.errChan
}

func (c *connection) onError(err error) {
	c.errChan <- err
}

func (c *connection) loop() {
	c.lg.Info("Start loop")
	var err error
	defer func() {
		c.cancel()
		if err != nil {
			c.onError(err)
		}
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}()

	for c.start.IsSet() {
		if err := c.dial(); err != nil {
			c.lg.Error("failed to dial", zap.Error(err))
			time.Sleep(time.Millisecond * 100)
			continue
		}
		ctx := c.ctx
		for {
			select {
			case <-ctx.Done():
				c.lg.Warn("context canceled")
				return
			case h := <-c.req:
				count := h.Size()
				if err = c.send(h); err != nil {
					c.lg.Error("Failed to send", zap.Error(err))
					return
				}
				for len(c.req) > 0 && count < defaultFlushSize {
					h = <-c.req
					if err = c.send(h); err != nil {
						c.lg.Error("Failed to send", zap.Error(err))
						return
					}
					count += h.Size()
				}
				if err = c.conn.Writer().Flush(); err != nil {
					c.lg.Error("Failed to flush buffer", zap.Error(err))
					return
				}
			case h := <-c.resp:
				if err = c.callback.onResponse(h.Uri, h.ErrCode, h.Message); err != nil {
					c.lg.Error("Failed to handle response", zap.Any("header", h))
					return
				}
			}
		}
	}
}

func (c *connection) dial() error {
	c.lg.Info("start dial")
	if conn, err := netpoll.DialConnection("tcp", c.edp, time.Second*5); err != nil {
		c.lg.Error("Failed to dial connection", zap.Error(err))
		return err
	} else {
		c.conn = conn
		_ = conn.SetOnRequest(c.onRequest)
		_ = conn.AddCloseCallback(c.onClose)
		c.lg.Info("connected")
	}
	return nil
}

// todo optimize performance
func (c *connection) send(h *Header) error {
	cLenS := 2
	if h.Size() >= byteThreshold {
		cLenS = 3
	}
	length, _ := TotalWithLength(h.Size() + 2 + 2 + cLenS)
	buffer, err := c.conn.Writer().Malloc(length)
	if err != nil {
		c.lg.Error("Failed to send", zap.Error(err))
		return err
	}
	body := make([]byte, h.Size()+2+2+cLenS)

	_, err = h.MarshalTo(body[2+2+cLenS:])
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint16(body[0:2], serviceSubproxy)
	binary.LittleEndian.PutUint16(body[2:4], UriCommonRequest)
	EncodeInt(h.Size(), body[4:4+cLenS])
	EncodeLen(body, buffer)
	c.lg.Debug("buffer", zap.ByteString("buffer", buffer), zap.Int("length", length), zap.Int("header size", h.Size()), zap.Int64("seqid", h.SeqId))
	return err
}

func (c *connection) onRequest(ctx context.Context, connection netpoll.Connection) error {
	if c.conn == nil || c.conn.LocalAddr() != connection.LocalAddr() {
		c.lg.Info("wrong conn", zap.String("local", connection.LocalAddr().String()))
		connection.Close()
		return nil
	}
	c.lg.Debug("request incoming", zap.String("local", connection.LocalAddr().String()))
	lenByte, err := c.conn.Reader().Peek(3)
	if err != nil {
		c.lg.Error("Failed to peek", zap.Error(err))
		return err
	}
	lenSize, length := DecodeInt(lenByte)
	buffer, err := c.conn.Reader().Next(length)
	if err != nil {
		c.lg.Error("Failed to read", zap.Error(err))
		return err
	}
	h := &Header{}
	// len(flex) + uri + service_id + content_length(flex)
	cLen, _ := DecodeInt(buffer[lenSize+2+2:])
	err = h.Unmarshal(buffer[lenSize+2+2+cLen:])
	if err != nil {
		c.lg.Error("unmarshal header error", zap.Error(err))
		return err
	}
	_ = c.conn.Reader().Release()
	c.lg.Debug("on request", zap.Int("length", len(buffer)), zap.Any("header", h))
	if IsEvent(h.Uri) {
		c.resp <- h
	} else {
		if value, ok := c.requests.Load(h.SeqId); ok {
			rc := value.(chan<- *Header)
			rc <- h
		} else {
			c.lg.Error("cannot find seqid", zap.Int64("seqid", h.SeqId))
		}
		return nil
	}
	return nil
}

func (c *connection) onClose(connection netpoll.Connection) error {
	if c.conn.LocalAddr() != connection.LocalAddr() {
		c.lg.Warn("Wrong connection", zap.String("wrong", connection.RemoteAddr().String()))
		return nil
	}

	c.lg.Info("Connection closed")
	c.onError(ERR_DISCONNECTED)
	c.cancel()
	c.requests.Range(func(key, value interface{}) bool {
		seqid := key.(int64)
		rc := value.(chan<- *Header)
		close(rc)
		c.lg.Error("disconnect", zap.Int64("seqid", seqid))
		c.requests.Delete(key)
		return true
	})
	return nil
}
