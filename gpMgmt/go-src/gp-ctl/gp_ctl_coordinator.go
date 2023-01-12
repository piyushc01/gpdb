package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "gp-ctl/agent"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var segListnerPort = 9000
var hostIP = "localhost"
var defaultDb = "template1"

func startInitsystem() error {
	// init main process
	addrs, err := net.InterfaceAddrs()
	iface, err := net.InterfaceByName("utun3")
	if err != nil {
		log.Fatalf("Error fetching ip address of coordinator")
		return err
	}
	addrs, err = iface.Addrs()
	hostIP = addrs[0].String()
	strArr := strings.Split(hostIP, "/")
	ipAddr := strArr[0]
	log.Println("Coordinator Address: " + ipAddr)

	// Get list of segments to be created
	segList := make(map[string][]segmentConfig)

	// Cluster Settings
	maxConnects := 150
	sharedBuff := "128000kB"
	// Cluster LC Settings
	lcSettings := lcStruct{"en_US.UTF-8", "en_US.UTF-8", "en_US.UTF-8",
		"en_US.UTF-8", "en_US.UTF-8", "en_US.UTF-8"}

	// Create Coordinator
	coordinator :=
		segmentConfig{"/tmp/1/qddir", maxConnects, sharedBuff, lcSettings,
			7000, 1, -1, hostIP, true}
	segList[hostIP] = append(segList["10.104.48.45"], coordinator)

	tempSeg := segmentConfig{"/tmp/1/datadir1", 150, "128000kB", lcSettings,
		7001, 2, 0, hostIP, false}
	segList[hostIP] = append(segList[hostIP], tempSeg)

	tempSeg = segmentConfig{"/tmp/1/datadir2", 150, "128000kB", lcSettings,
		7002, 3, 1, hostIP, false}
	segList[hostIP] = append(segList[hostIP], tempSeg)

	tempSeg = segmentConfig{"/tmp/1/datadir3", 150, "128000kB", lcSettings,
		7003, 4, 2, hostIP, false}
	segList[hostIP] = append(segList[hostIP], tempSeg)

	tempSeg = segmentConfig{"/tmp/1/datadir4", 150, "128000kB", lcSettings,
		7004, 5, 3, hostIP, false}
	segList[hostIP] = append(segList[hostIP], tempSeg)

	// Start listener on segment host if not running
	for host := range segList {
		// Call a dummy RPC on the port
		retVal := testConnectSegment(host, segListnerPort)
		if retVal {
			//listener is running
			log.Println("Listener running on host" + host)
		} else {
			//TODO Run listener in daemon mode
			//If not running
			// exec ( ssh hostname "gp_ctl init listener --port 9000  )
		}
		connectAndCreateSegments(segList[host], ipAddr)
	}
	for host := range segList {
		for _, seg := range segList[host] {
			err = registerSegment(seg, ipAddr, coordinator.portNum)
			if err != nil {
				log.Fatalf("Error while creting segment:" + err.Error())
			}
		}
	}
	//err = registerSegment(coordinator, ipAddr, coordinator.portNum)

	// Create per host list of segments and call createAndConfigSegments

	// Get list of segments created and started
	// Register segments
	// Restart cluster
	return nil
}

func registerSegment(seg segmentConfig, ipAddr string, coorinatorPort int) error {
	//TODO redefined here, move this to init
	gphome := os.Getenv("GPHOME")
	if gphome == "" {
		fmt.Println("GPHOME is not set. Exiting.")
		err := errors.New("GPHOME is not set, exiting")
		return err
	}
	os.Setenv("PGPORT", strconv.Itoa(coorinatorPort))
	conn, err := ConnectToSegmentDB(defaultDb, ipAddr, strconv.Itoa(coorinatorPort))
	if err != nil {
		log.Fatalf("Error connecting coordinator DB")
		return err
	}
	defer conn.Close()
	var queryStr string
	//queryStr = "SELECT pg_catalog.gp_add_segment(0::int2, -1::int2, 'p', 'p', 's', 'u', 7000, '10.104.50.186', '10.104.50.186', '/tmp/1/qddir')"
	queryStr = fmt.Sprintf("SELECT pg_catalog.gp_add_segment(%d::int2, %d::int2, 'p', 'p', 's', 'u', %d, '%s', '%s', '%s')",
		seg.dbid, seg.contentId, seg.portNum, ipAddr, ipAddr, seg.dataDir)
	_, err = conn.Query(queryStr)
	if err != nil {
		fmt.Println("ERROR registering coordinator", err.Error())
		log.Fatalf("ERROR registering coordinator", err.Error())
		return err
	}

	return nil
}

// Returns false if connection fails
func testConnectSegment(hostName string, potNum int) bool {

	connStr := hostName + ":" + strconv.Itoa(potNum)
	conn, err := grpc.Dial(connStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func connectAndCreateSegments(segList []segmentConfig, hostName string) {
	connStr := hostName + ":" + strconv.Itoa(segListnerPort)
	conn, err := grpc.Dial(connStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewCliToSegmentHostClient(conn)
	var req []*pb.SegmentConfig
	for _, seg := range segList {
		tempReq := &pb.SegmentConfig{
			DataDir:        seg.dataDir,
			MaxConnections: int32(seg.maxConnections),
			SharedBuffers:  seg.sharedBuffers,
			LcConfig: &pb.LcStruct{
				LcCollate:  seg.lcConfig.lcCollate,
				LcCtpye:    seg.lcConfig.lcCtpye,
				LcMessages: seg.lcConfig.lcMessages,
				LcMonetary: seg.lcConfig.lcMonetary,
				LcNumeric:  seg.lcConfig.lcNumeric,
				LcTime:     seg.lcConfig.lcTime,
			},
			PortNum:       int32(seg.portNum),
			Dbid:          int32(seg.dbid),
			ContentId:     int32(seg.contentId),
			CoordinatorIP: seg.coordinatorIP,
		}
		req = append(req, tempReq)
	}

	// add data to it
	/*
		req := &pb.SegmentConfigList{
			SegConfig: []*pb.SegmentConfig{{
				DataDir:        "/tmp/datadir1",
				MaxConnections: 8,
				SharedBuffers:  "sharedbuffer",
				LcConfig: &pb.LcStruct{
					LcCollate:  "C",
					LcCtpye:    "en_US.UTF-8",
					LcMessages: "C",
					LcMonetary: "C",
					LcNumeric:  "C",
					LcTime:     "C",
				},
				PortNum:       7010,
				Dbid:          9,
				ContentId:     3,
				CoordinatorIP: "0.0.0.0",
			},
			},
		} */

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	request := &pb.SegmentConfigList{
		SegConfig: req,
	}
	r, err := c.CreateAndConfigureSegments(context.Background(), request)
	if err != nil {
		log.Fatalf("could not execute: %v", err)
	}
	log.Printf("Output: %s", r.SegStatus)
}
