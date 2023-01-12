package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type lcStruct struct {
	lcCollate  string
	lcCtpye    string
	lcMessages string
	lcMonetary string
	lcNumeric  string
	lcTime     string
}

type segmentConfig struct {
	dataDir        string
	maxConnections int
	sharedBuffers  string
	lcConfig       lcStruct
	portNum        int
	dbid           int
	contentId      int
	coordinatorIP  string
	isCoordinator  bool
}

type initdbStatus struct {
	dataDir string
	output  string
	err     error
}

var wg sync.WaitGroup

func main() {
	fmt.Println("GP-CT starting")
	// sample: gp_ctl init
	// sample: gp_ctl init --listener --port 9000
	initFlag := flag.Bool("init", false, "Starts initialization of gpdb")
	listenerFlag := flag.Bool("listener", false, "Starts segment listener process")
	//listnerPort := flag.Int("port", segListnerPort, "Listen port fo initsystem segment listner")
	flag.Parse()
	if *initFlag {
		//Check if segment listener mode, start listener
		if *listenerFlag {
			fmt.Println("GP-CT starting listener")
			// Run as a listener to accept segment creation request
			startSegmnetListener(segListnerPort)
			return
		}
		//GP-CTL init main
		if !*listenerFlag {
			fmt.Println("GP-CT starting initsystem")
			err := startInitsystem()
			if err != nil {
				fmt.Println("GP-CT finished with error")
			}
			// createGpCluster()

		}

	}

	if !*initFlag && !*listenerFlag {
		fmt.Println("GP-CT no parameters provided, existing")
		return
	}
}

func createGpCluster() {
	var seglist []segmentConfig

	// Cluster Settings
	maxConnects := 150
	sharedBuff := "128000kB"
	// Cluster LC Settings
	lcSettings := lcStruct{"en_US.UTF-8", "en_US.UTF-8", "en_US.UTF-8",
		"en_US.UTF-8", "en_US.UTF-8", "en_US.UTF-8"}

	// Create Coordinator
	coordinator :=
		segmentConfig{"/tmp/1/qddir", maxConnects, sharedBuff, lcSettings,
			7000, 1, 0, "10.104.48.45/32", true}
	seglist = append(seglist, coordinator)

	seglist = append(seglist,
		segmentConfig{"/tmp/1/datadir1", 150, "128000kB", lcSettings,
			7001, 2, 1, "10.104.48.45/32", false})
	seglist = append(seglist,
		segmentConfig{"/tmp/1/datadir2", 150, "128000kB", lcSettings,
			7002, 3, 2, "10.104.48.45/32", false})
	seglist = append(seglist,
		segmentConfig{"/tmp/1/datadir3", 150, "128000kB", lcSettings,
			7003, 4, 3, "10.104.48.45/32", false})
	seglist = append(seglist,
		segmentConfig{"/tmp/1/datadir4", 150, "128000kB", lcSettings,
			7004, 5, 4, "10.104.48.45/32", false})

	var messages = make(chan initdbStatus, len(seglist))
	// Create segments in parallel
	for _, seg := range seglist {
		wg.Add(1)
		go createPgInstance(seg, messages)
	}

	// wait for all to finish and collect status
	fmt.Println("Waiting for go-routines to finish...")
	wg.Wait()
	close(messages)
	for out := range messages {
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
}

func createPgInstance(segConfig segmentConfig, channel chan initdbStatus) {
	defer wg.Done()
	// Check if data-dir is empty.
	if checkIfDirEmpty(segConfig.dataDir) == false {
		fmt.Println("ERROR: Segment directory is not empty. Can't create segment")
		err := errors.New("segment directory is not empty, not able to create segment")
		channel <- initdbStatus{err: err, output: "", dataDir: segConfig.dataDir}
		return
	}
	// TODO: add this to init function
	gphome := os.Getenv("GPHOME")
	if gphome == "" {
		fmt.Println("GPHOME is not set. Exiting.")
		err := errors.New("GPHOME is not set, exiting")
		channel <- initdbStatus{err: err, output: "", dataDir: segConfig.dataDir}
		return
	}
	initdb := filepath.Join(gphome, "bin", "initdb")
	var gpinitCmd = exec.Command(initdb,
		"-D", segConfig.dataDir, "-E", "UNICODE", "--max_connections="+strconv.Itoa(segConfig.maxConnections),
		"--shared_buffers="+segConfig.sharedBuffers, "--data-checksums", "--lc-collate="+segConfig.lcConfig.lcCollate,
		"--lc-ctype="+segConfig.lcConfig.lcCtpye, "--lc-messages="+segConfig.lcConfig.lcMessages,
		"--lc-monetary="+segConfig.lcConfig.lcMonetary, "--lc-numeric="+segConfig.lcConfig.lcNumeric,
		"--lc-time="+segConfig.lcConfig.lcTime)
	out, retVal := gpinitCmd.Output()
	if retVal != nil {
		fmt.Println("ERROR executing initdb")
		fmt.Println(retVal)
		segStatus := initdbStatus{dataDir: segConfig.dataDir, output: string(out[:]), err: retVal}
		channel <- segStatus
		return
	}
	segStatus := initdbStatus{dataDir: segConfig.dataDir, output: string(out[:]), err: nil}

	updateFailed, err := updatePgConfEntries(segConfig)
	if updateFailed == true {
		segStatus.err = err
		channel <- segStatus
	}
	// Update pg_hba.conf
	updateFailed, err = updatePgHbaConfEntries(segConfig)
	if updateFailed == true {
		segStatus.err = err
		channel <- segStatus
	}
	// TODO: Additional pg-config

	// Start all segments in admin mode
	pgCtlPath := filepath.Join(gphome, "bin", "pg_ctl")
	//export PGPORT=$GP_PORT;$PG_CTL -w -l $GP_DIR/log/startup.log -D $GP_DIR -o "-i -p $GP_PORT -c gp_role=utility -m" start >> ${LOG_FILE} 2>&1
	// pg_ctl  -w -l /tmp/1/datadir2/log/startup.log -D /tmp/1/datadir2 -o "-i -p 7002 -c gp_role=utility -m" start
	startCmd := exec.Command(pgCtlPath, "-w", "-l", filepath.Join(segConfig.dataDir),
		"-o", "\"-i -p "+strconv.Itoa(segConfig.portNum)+" -c gp_role=utility -m\"", "start")
	startCmd.Env = os.Environ()
	startCmd.Env = append(startCmd.Env, "PGPORT="+strconv.Itoa(segConfig.portNum))

	startCmd = exec.Command(pgCtlPath, "start", "-l",
		filepath.Join(segConfig.dataDir, "log", "startup.log"),
		"-D", segConfig.dataDir, "-o", "-i -p "+strconv.Itoa(segConfig.portNum)+" -c gp_role=utility -m")
	startCmd.Env = os.Environ()
	startCmd.Env = append(startCmd.Env, "PGPORT="+strconv.Itoa(segConfig.portNum))

	out, retVal = startCmd.Output()
	if retVal != nil {
		fmt.Println("Error starting segment:" + segConfig.dataDir)
		fmt.Println(retVal.Error())
	}

	channel <- segStatus

}
func updatePgHbaConfEntries(segConfig segmentConfig) (bool, error) {
	// Update pg-hba.conf
	pgHbaPath := filepath.Join(segConfig.dataDir, "pg_hba.conf")
	f, err := os.OpenFile(pgHbaPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error while opening pg_hba.conf file:" + pgHbaPath)
		fmt.Println(err.Error())
		return true, err
	}
	// Add entry for each coordinator IP address in CIDR format :host all all <coordinator-ip-address> trust
	iplist := strings.Split(segConfig.coordinatorIP, ",")
	for _, cordIP := range iplist {
		var hbaStr = "host all all " + cordIP + " trust"
		if _, err = f.Write([]byte("\n" + hbaStr + "\n")); err != nil {
			fmt.Println("Error while writing the pg_hba file:" + pgHbaPath)
			fmt.Println(err.Error())
			return true, err
		}
	}
	// Get current username
	user, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user while updating pghba.conf.")
		return true, err
	}
	username := user.Username
	// Get all local ip addresses (IPV4 and IPv6) and update to pg_hba.conf
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		//fmt.Println("IPv4: ", addr)
		if addr.String() == "127.0.0.1/8" || addr.String() == "::1/128" || addr.String() == "fe80::1/64" {
			continue
		}
		hbaStr := "host all " + username + " " + addr.String() + " trust"
		if _, err = f.Write([]byte("\n" + hbaStr + "\n")); err != nil {
			fmt.Println("Error while writing the pg_hba file:" + pgHbaPath)
			fmt.Println(err.Error())
			return true, err
		}
	}

	// Close pg_hba.conf file
	if err := f.Close(); err != nil {
		fmt.Println("Error while closing pg_hba.conf file:" + pgHbaPath)
		fmt.Println(err)
		return true, err
	}
	return false, nil
}

// updatePgConfEntries returns true on failure and returns error
func updatePgConfEntries(segConfig segmentConfig) (bool, error) {
	// Assign port number in PG-CONF: port=<port-no>
	var portStr = "port=" + strconv.Itoa(segConfig.portNum)
	// Update listen address in PG-CONF : listen_addresses='*'
	var listenAddressStr = "listen_addresses='*'"
	var pgconfPath = filepath.Join(segConfig.dataDir, "postgresql.conf")
	// Update content ID: gp_contentid=<Content-ID>
	var contentIdStr = "gp_contentid=" + strconv.Itoa(segConfig.contentId)
	// Update dbid : gp_dbid=<dbid>
	var dbidStr = "gp_dbid=" + strconv.Itoa(segConfig.dbid)
	f, err := os.OpenFile(pgconfPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error while opening postgres.conf file:" + pgconfPath)
		fmt.Println(err.Error())
		return true, err
	}
	if _, err = f.Write([]byte("\n# GPDB Config\n" + portStr + "\n" + listenAddressStr + "\n" + contentIdStr + "\n" + dbidStr + "\n")); err != nil {
		fmt.Println("Error while writing the postgres file:" + pgconfPath)
		fmt.Println(err.Error())
		return true, err
	}
	// Close config file
	if err := f.Close(); err != nil {
		fmt.Println("Error while closing postgres.conf file:" + pgconfPath)
		fmt.Println(err)
		return true, err
	}
	return false, nil
}

// checkIfDirEmpty returns true if directory does not exist or is empty
func checkIfDirEmpty(dataDir string) bool {
	var f, err = os.Open(dataDir)
	if err != nil {
		// This means directory does not exist
		return true
	}
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		fmt.Println("ERROR: Directory does not exists")
		return true
	}
	return false
}
