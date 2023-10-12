package main

import (
	context "context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	sync "sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	status "google.golang.org/grpc/status"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
// если хочется, то для красоты можно разнести логику по разным файликам

type BizServis struct {
	UnimplementedBizServer
	UnimplementedAdminServer

	stat Stater
	acl  map[string][]string
	ml   myLogger
}

func NewBizServer(acl map[string][]string) *BizServis {
	bs := &BizServis{acl: acl}
	bs.stat.Init()
	bs.ml.Init()
	return bs
}

func StartMyMicroservice(ctx context.Context, addr, aclData string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("cant listen on port : %w", err)
	}
	m := make(map[string][]string)
	err = json.Unmarshal([]byte(aclData), &m)
	if err != nil {
		lis.Close()
		return fmt.Errorf("%w : StartMyMicroservice", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
		grpc.StreamInterceptor(authStreamInterceptor),
	)
	bizServ := NewBizServer(m)
	RegisterAdminServer(server, bizServ)
	RegisterBizServer(server, bizServ)
	//nolint:errcheck
	go server.Serve(lis)
	go func() {
		for {
			<-ctx.Done()
			server.Stop()
			return
		}
	}()
	return nil
}

func authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	if len(md.Get("consumer")) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	consumer := md.Get("consumer")[0]

	serv, ok := info.Server.(*BizServis)
	if !ok && serv == nil {
		return nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	if valid(serv.acl, consumer, info.FullMethod) {
		event := &Event{Consumer: consumer, Method: info.FullMethod}
		p, ok := peer.FromContext(ctx)
		if ok {
			event.Host = p.Addr.String()
		}
		serv.ml.PrintLogsToAll(event)
		serv.stat.MakeStat(info.FullMethod, consumer)
		return handler(ctx, req)
	}
	return nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())

}

func authStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	if len(md.Get("consumer")) == 0 {
		return status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}
	consumer := md.Get("consumer")[0]

	serv, ok := srv.(*BizServis)
	if !ok && serv == nil {
		return status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	if valid(serv.acl, consumer, info.FullMethod) {
		event := &Event{Consumer: consumer, Method: info.FullMethod}
		p, ok := peer.FromContext(ss.Context())
		if ok {
			event.Host = p.Addr.String()
		}
		serv.ml.PrintLogsToAll(event)
		serv.stat.MakeStat(info.FullMethod, consumer)
		return handler(srv, ss)

	}
	return status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
}

func valid(m map[string][]string, consumer, method string) bool {
	for _, v := range m[consumer] {
		if len(v) > 0 && v[len(v)-1:] == "*" {
			if strings.HasPrefix(method, v[:len(v)-1]) {
				return true
			}
		}
		if v == method {
			return true
		}
	}
	return false
}

type myLogger struct {
	logs map[int64]chan *Event
	mu   *sync.RWMutex
	id   int64
}

func (ml *myLogger) Init() {
	ml.id = 0
	ml.mu = &sync.RWMutex{}
	ml.logs = make(map[int64]chan *Event)
}

func (ml *myLogger) NewLoger() (chan *Event, int64) {
	ml.mu.Lock()
	ch := make(chan *Event)
	ml.logs[ml.id] = ch
	ml.id++
	ml.mu.Unlock()
	return ch, ml.id - 1
}

func (ml *myLogger) DeleteLoger(i int64) {
	ml.mu.Lock()
	ch, ok := ml.logs[i]
	if !ok {
		return
	}
	close(ch)
	delete(ml.logs, i)
	ml.mu.Unlock()
}

func (ml *myLogger) PrintLogsToAll(event *Event) {
	ml.mu.RLock()
	for _, v := range ml.logs {
		v <- event
	}
	ml.mu.RUnlock()
}

func (s *BizServis) Logging(non *Nothing, streams Admin_LoggingServer) error {
	ch, id := s.ml.NewLoger()
	var event *Event
	for {
		select {
		case <-streams.Context().Done():
			s.ml.DeleteLoger(id)
			return nil
		case event = <-ch:
			e := streams.Send(event)
			if e != nil {
				s.ml.DeleteLoger(id)
				return e
			}
		}
	}
}

type Stater struct {
	id             uint64
	StatFirstTry   bool
	statByConsumer map[uint64]map[string]uint64
	statByMethod   map[uint64]map[string]uint64
	mu             *sync.Mutex
}

func (s *Stater) Init() {
	s.id = 0
	s.mu = &sync.Mutex{}
	s.statByConsumer = make(map[uint64]map[string]uint64)
	s.statByMethod = make(map[uint64]map[string]uint64)
	s.StatFirstTry = true
}

func (s *Stater) AddListener() uint64 {
	s.mu.Lock()
	s.id++
	s.statByConsumer[s.id] = make(map[string]uint64)
	s.statByMethod[s.id] = make(map[string]uint64)
	defer s.mu.Unlock()
	return s.id
}

func (s *Stater) RemoveListener(id uint64) {
	s.mu.Lock()
	delete(s.statByConsumer, id)
	delete(s.statByMethod, id)
	s.mu.Unlock()
}

func (s *Stater) Stat(id uint64) (map[string]uint64, map[string]uint64) {
	s.mu.Lock()
	sbc := s.statByConsumer[id]
	sbm := s.statByMethod[id]
	sbcRet := make(map[string]uint64)
	sbmRet := make(map[string]uint64)
	for k, v := range sbc {
		sbcRet[k] = v
	}
	for k, v := range sbm {
		sbmRet[k] = v
	}
	s.statByConsumer[id] = make(map[string]uint64)
	s.statByMethod[id] = make(map[string]uint64)
	s.mu.Unlock()
	return sbcRet, sbmRet
}

func (s *Stater) MakeStat(method, consumer string) {
	s.mu.Lock()
	for _, v := range s.statByConsumer {
		v[consumer]++
	}
	for _, v := range s.statByMethod {
		v[method]++
	}
	if method == "/main.Admin/Statistics" && s.StatFirstTry {
		s.StatFirstTry = false
		for _, v := range s.statByConsumer {
			v[consumer]--
		}
		for _, v := range s.statByMethod {
			v[method]--
		}
	}
	s.mu.Unlock()
}

func (s *BizServis) Statistics(statInterval *StatInterval, streams Admin_StatisticsServer) error {
	id := s.stat.AddListener()
	for {
		select {
		case <-streams.Context().Done():
			s.stat.RemoveListener(id)
			return nil
		case <-time.After(time.Duration(statInterval.IntervalSeconds) * time.Second):
			sbc, sbm := s.stat.Stat(id)
			st := &Stat{
				ByConsumer: sbc,
				ByMethod:   sbm,
			}
			err := streams.Send(st)
			if err != nil {
				return err
			}
		}
	}
}

func (s *BizServis) Check(ctx context.Context, non *Nothing) (*Nothing, error) {
	fmt.Println("Check")
	return non, nil
}
func (s *BizServis) Add(ctx context.Context, non *Nothing) (*Nothing, error) {
	fmt.Println("Add")
	return non, nil
}
func (s *BizServis) Test(ctx context.Context, non *Nothing) (*Nothing, error) {
	fmt.Println("Test")
	return non, nil
}
