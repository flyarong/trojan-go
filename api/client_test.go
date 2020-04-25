package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
)

func TestClientAPI(t *testing.T) {
	meter := &stat.MemoryTrafficMeter{}
	go RunClientAPIService(context.Background(), &conf.GlobalConfig{
		API: conf.APIConfig{
			APIAddress: common.NewAddress("127.0.0.1", 10000, "tcp"),
		},
	}, meter)
	meter.Count("test", 123, 456)
	time.Sleep(time.Second)
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	client := NewTrojanServiceClient(conn)
	reply, err := client.QueryStats(context.Background(), &StatsRequest{})
	common.Must(err)
	fmt.Println(reply.DownloadTraffic, reply.UploadTraffic)
	if reply.DownloadTraffic != 456 || reply.UploadTraffic != 123 {
		t.Fatal("wrong result")
	}
}

func TestRealClientAPI(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	client := NewTrojanServiceClient(conn)
	reply, err := client.QueryStats(context.Background(), &StatsRequest{})
	common.Must(err)
	fmt.Println(reply.DownloadTraffic, reply.UploadTraffic)
}
