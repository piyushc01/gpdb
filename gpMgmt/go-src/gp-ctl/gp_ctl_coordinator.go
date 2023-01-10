package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func registerMasterSegment(seg segmentConfig) error {
	//TODO redefined here, move this to init
	gphome := os.Getenv("GPHOME")
	if gphome == "" {
		fmt.Println("GPHOME is not set. Exiting.")
		err := errors.New("GPHOME is not set, exiting")
		return err
	}

	psqlPath := filepath.Join(gphome, "bin", "pg_ctl")
	// Create DB: $PSQL -p $GP_PORT -d "$DEFAULTDB" -c"create database \"${DATABASE_NAME}\";" >> $LOG_FILE 2>&1
	cmd := exec.Command(psqlPath, "-p", strconv.Itoa(seg.portNum), "-d", seg.dataDir, "-X", "-A", "-t",
		"-c", "SELECT count(*) FROM gp_segment_configuration WHERE content="+strconv.Itoa(seg.contentId)+
			" AND preferred_role=p;")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PGOPTIONS=\"-c gp_role=utility\"")
	out, retVal := cmd.Output()
	if retVal != nil || strings.Trim(string(out[:]), " \n\t") != "0" {
		fmt.Println("ERROR: Segment alreadt registered with DBID:" + strconv.Itoa(seg.dbid))

	}
	return nil
}
