package connection_pool

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type idleConn struct {
	c              net.Conn
	lastActiveTime time.Time // 上次活跃时间，例如被创建、被使用，记录上次 TCP 的使用时间
}

type connRequest struct {
	conn chan net.Conn
}

type Pool struct {
	idleConns    chan *idleConn           // 空闲连接 chan
	requestQueue []connRequest            // 阻塞的请求
	maxCount     int                      // 最大连接数
	currCount    int32                    // 当前连接数
	maxIdleCount int                      // 最大空闲连接数
	maxIdleTime  time.Duration            // 最大空闲连接时间
	initCount    int                      // 初始连接数
	factory      func() (net.Conn, error) // 工厂模式
	lock         sync.Mutex
}

func NewPool(
	maxCount int,
	maxIdleCount int,
	maxIdleTime time.Duration,
	initCount int,
	factory func() (net.Conn, error),
) (*Pool, error) {
	// 当 maxIdleCount > maxCount
	//	 也就是说最大空闲连接数大于最大连接数，此时应该 增加 最大连接数，还是减少最大空闲连接数？
	//   应该选择后者，强制限制以最大连接数为主，维护系统安全
	if maxIdleCount > maxCount {
		maxIdleCount = maxCount
	}

	// 同样的，初始连接个数也不应该大于最大空闲连接数
	if initCount > maxIdleCount {
		initCount = maxIdleCount
	}

	// 初始化连接
	idleConns := make(chan *idleConn, maxIdleCount)
	for i := 0; i < initCount; i++ {
		c, err := factory()
		if err != nil {
			return nil, err
		}
		idleConns <- &idleConn{c: c, lastActiveTime: time.Now()}
	}

	return &Pool{
		idleConns:    idleConns,
		maxCount:     maxCount,
		maxIdleCount: maxIdleCount,
		maxIdleTime:  maxIdleTime,
		factory:      factory,
	}, nil
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	for {
		select {
		case c := <-p.idleConns:
			// 获取到连接
			if c.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				// 连接已经过期
				_ = c.c.Close()
				atomic.AddInt32(&p.currCount, -1)
				continue
			}
			return c.c, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// 没有空闲连接，也没有超时
			p.lock.Lock()
			// 如果当前连接数大于等于最大连接数
			if int(p.currCount) >= p.maxCount {
				// 进入等待队列
				req := connRequest{conn: make(chan net.Conn)}
				p.requestQueue = append(p.requestQueue, req)
				p.lock.Unlock()

				select {
				case c := <-req.conn:
					return c, nil
				case <-ctx.Done():
					// 如果超时，还需要销毁自己的 chan，选择保险的方法是放回去
					go func() {
						c := <-req.conn
						_ = p.Put(ctx, c)
					}()
					return nil, ctx.Err()
				}
			}

			p.lock.Unlock()

			conn, err := p.factory()
			if err != nil {
				return nil, err
			}
			atomic.AddInt32(&p.currCount, 1)

			return conn, nil
		}
	}
}

func (p *Pool) Put(ctx context.Context, conn net.Conn) error {
	p.lock.Lock()

	if len(p.requestQueue) != 0 {
		// 拿取队首的，容易超时，但是遵循先进先出
		// 拿取队尾，基本上不超时，但是不易理解
		req := p.requestQueue[0]
		p.requestQueue = p.requestQueue[1:]
		p.lock.Unlock()

		req.conn <- conn
		return nil
	}
	p.lock.Lock()

	// 没有阻塞的请求
	select {
	// 将连接放入连接池中
	case p.idleConns <- &idleConn{c: conn, lastActiveTime: time.Now()}:
	default:
		atomic.AddInt32(&p.currCount, -1)
		return conn.Close()
	}

	return nil
}
