package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "gp-ctl/agent"
	"log"
	"net"
	"strconv"
)

type server struct {
	pb.CliToSegmentHostServer
}

func (s *server) CreateAndConfigureSegments(ctx context.Context, req *pb.SegmentConfigList) (*pb.SegmentStatusList, error) {
	log.Printf("Recieved configuration: %v", req.SegConfig)
	var messages = make(chan initdbStatus, len(req.SegConfig))
	// Create segments in parallel
	for _, seg := range req.SegConfig {
		tempLcConfig := lcStruct{seg.LcConfig.LcCollate, seg.LcConfig.LcCtpye, seg.LcConfig.LcMessages,
			seg.LcConfig.LcMonetary, seg.LcConfig.LcNumeric, seg.LcConfig.LcTime}
		tempSegConfig := segmentConfig{seg.DataDir, int(seg.MaxConnections), seg.SharedBuffers,
			tempLcConfig, int(seg.PortNum), int(seg.Dbid), int(seg.ContentId), seg.CoordinatorIP, false}
		wg.Add(1)
		go createPgInstance(tempSegConfig, messages)
	}
	// wait for all to finish and collect status
	fmt.Println("Waiting for go-routines to finish...")
	wg.Wait()
	close(messages)
	var segStatus []*pb.SegStatus
	for out := range messages {
		retVal := 0
		if out.err != nil {
			retVal = 1
		}
		res := &pb.SegStatus{
			DataDir:    out.dataDir,
			RetVal:     int32(retVal),
			ErrMessage: out.output,
		}
		segStatus = append(segStatus, res)

		if out.err == nil {
			fmt.Println("Segment Create Success:" + out.dataDir)
			continue
		}
		if out.err != nil {
			fmt.Println("Segment creation failed:" + out.dataDir)
			fmt.Println("Error:" + out.err.Error())
			continue
		}
	}
	res := &pb.SegmentStatusList{
		SegStatus: segStatus,
	}

	//Result: req.A + req.B,

	return res, nil
}

//// create initdb command and execute it
//out, err := exec.Command(in.Cmd, in.Args...).Output()
//if err != nil {
//	log.Fatal(err)
//}
//return &pb.CommandResponse{Body: string(out)}, nil

func startSegmnetListener(port int) {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCliToSegmentHostServer(grpcServer, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
