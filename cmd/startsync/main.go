package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/uluyol/startsync/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const numAttempt = 20

func run() error {
	if len(os.Args) < 5 {
		return fmt.Errorf("usage: %s [startsync addr] [key] [waitcount] cmd args...", os.Args[0])
	}
	addr := os.Args[1]
	key := os.Args[2]
	count64, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil || count64 <= 0 {
		return fmt.Errorf("invalid count: %v", err)
	}
	count := int32(count64)
	cmdline := os.Args[4:]
	gclient, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewStartSyncClient(gclient)
	req := pb.WaitReq{Key: key, Count: count}
	var resp *pb.WaitResp
	for i := 0; i < numAttempt; i++ {
		resp, err = client.Wait(context.Background(), &req)
		if err == nil {
			break
		}
	}
	gclient.Close()
	if err != nil {
		return err
	}
	if resp.Count != count {
		return fmt.Errorf("sync server registered %d nodes to wait for even though %d was requested", resp.Count, count)
	}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
