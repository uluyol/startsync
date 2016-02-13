package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/golang/glog"
	"github.com/uluyol/startsync/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type counter struct {
	has int32
	req int32
	ann chan struct{}
}

func syncServer() chan<- ssInput {
	in := make(chan ssInput)
	go func() {
		counters := make(map[string]counter)
		for {
			input := <-in
			if !input.arrive {
				if c, ok := counters[input.key]; ok {
					c.has--
					counters[input.key] = c
				}
				continue
			}
			var c counter
			if t, ok := counters[input.key]; ok {
				c = t
			} else {
				c.req = input.count
				c.ann = make(chan struct{})
			}
			c.has++
			input.respChan <- ssResp{done: c.ann, count: c.req}
			if c.has == c.req {
				glog.Infof("Got all %d nodes for %q", c.has, input.key)
				delete(counters, input.key)
				close(c.ann)
			} else {
				counters[input.key] = c
			}
		}
	}()
	return in
}

type ssResp struct {
	done  <-chan struct{}
	count int32
}

type ssInput struct {
	respChan chan<- ssResp
	key      string
	count    int32
	arrive   bool
}

type rpcServer chan<- ssInput

func (s rpcServer) Wait(ctx context.Context, req *pb.WaitReq) (*pb.WaitResp, error) {
	glog.Infof("Got request to wait for %d nodes for %q", req.Count, req.Key)
	rc := make(chan ssResp)
	s <- ssInput{
		respChan: rc,
		key:      req.Key,
		count:    req.Count,
		arrive:   true,
	}
	got := <-rc
	resp := pb.WaitResp{Count: got.count}
	select {
	case <-got.done:
		resp.Start = true
		return &resp, nil
	case <-ctx.Done():
		glog.Infof("Connection ended for %q", req.Key)
		s <- ssInput{
			respChan: nil,
			key:      req.Key,
			count:    req.Count,
			arrive:   false,
		}
		return nil, ctx.Err()
	}
}

var (
	portNum = flag.Int("p", 6080, "port to listen on")
)

func main() {
	flag.Parse()
	if *portNum <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	ssInputChan := syncServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *portNum))
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterStartSyncServer(gs, rpcServer(ssInputChan))
	gs.Serve(lis)
}
