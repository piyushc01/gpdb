package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Following are all the map names for the various connection options
const (
	PGUSER      = "user"
	PGPASS      = "password"
	PGHOST      = "host"
	PGPORT      = "port"
	DBNAME      = "dbname"
	SSL         = "sslmode"
	UTILITYMODE = "utility"
	// TODO: Add the other connection options from https://www.postgresql.org/docs/8.3/static/libpq-connect.html
)

var dbConn *sql.DB

func StartGPDB() error {
	cmdStr, err := exec.LookPath("gpstart")
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdStr, "-a")
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}

func StopGPDB() error {
	cmdStr, err := exec.LookPath("gpstop")
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdStr, "-af")
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getOSEnv(env string, def string) string {
	if s := os.Getenv(env); s != "" {
		return s
	}
	return def
}

func getDefaultPGUser() string {
	return getOSEnv("PGUSER", os.Getenv("USER"))
}

func getDefaultPGPort() string {
	return getOSEnv("PGPORT", "5432")
}

func getDefaultPGDBName() string {
	return getOSEnv("PGDATABASE", os.Getenv("USER"))
}

func getDefaultPGHost() string {
	return getOSEnv("PGHOST", "localhost")
}

func getDefaultPGSSLMode() string {
	return getOSEnv("SSLMODE", "disable")
}

func GetDefaultPGOptions() map[string]string {
	return map[string]string{
		"user":    getDefaultPGUser(),
		"host":    getDefaultPGHost(),
		"port":    getDefaultPGPort(),
		"dbname":  getDefaultPGDBName(),
		"sslmode": getDefaultPGSSLMode(),
	}

}

// GPDBConnect connects to the named database
func GPDBConnect(dbname string) (*sql.DB, error) { // Connect to the given dname on default ports
	opts := GetDefaultPGOptions()
	if dbname != "" {
		opts[DBNAME] = dbname
	}
	return GPDBConnectWithOptions(opts, false)
}

func GPDBConnectWithOptions(opts map[string]string, utility bool) (*sql.DB, error) {
	var dbcon *sql.DB
	log.Println("Connecting to gpdb with the following options: %v", opts)

	// Disable Kerberos ENV variables
	err := DisableKerberos()
	if err != nil {
		return dbcon, err
	}

	optArray := make([]string, 0, len(opts)+1)
	for k, v := range opts {
		optArray = append(optArray, k+"="+v)
	}
	if utility {
		optArray = append(optArray, "options='-c gp_role=utility'")
	}
	return sql.Open("postgres", strings.Join(optArray, " "))
}

func ConnectToSegmentDB(dbname string, host string, port string) (*sql.DB, error) {
	opts := map[string]string{
		"user":    getDefaultPGUser(),
		"host":    host,
		"port":    port,
		"dbname":  dbname,
		"sslmode": getDefaultPGSSLMode(),
	}
	return GPDBConnectWithOptions(opts, true)
}

type RowIter struct {
	colNames    []string
	rows        *sql.Rows
	readCols    []interface{}
	nullStrings []sql.NullString
}

func NewRowIter(r *sql.Rows) (*RowIter, error) {
	cn, err := r.Columns()
	if err != nil {
		return nil, err
	}
	ri := RowIter{
		colNames:    cn,
		rows:        r,
		readCols:    make([]interface{}, len(cn)),
		nullStrings: make([]sql.NullString, len(cn)),
	}
	for i := range cn {
		ri.readCols[i] = &(ri.nullStrings[i])
	}
	return &ri, nil
}

func (r *RowIter) Next() bool {
	return r.rows.Next()
}

func (r *RowIter) GetNextRow() ([]string, error) {
	if err := r.rows.Scan(r.readCols...); err != nil {
		return []string{}, err
	}
	ret := make([]string, len(r.colNames))
	for i := range ret {
		ret[i] = r.nullStrings[i].String
	}
	return ret, nil
}

func (r *RowIter) Columns() []string {
	return r.colNames
}

func GetResultsArray(rows *sql.Rows) ([][]string, error) {
	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	readCols := make([]interface{}, len(colNames))
	nullStrings := make([]sql.NullString, len(colNames))
	for i := range readCols {
		readCols[i] = &nullStrings[i]
	}
	ret := [][]string{}
	for rows.Next() {
		if rErr := rows.Scan(readCols...); rErr != nil {
			return nil, rErr
		}
		dst := make([]string, len(colNames))
		for i := range nullStrings {
			dst[i] = nullStrings[i].String
		}
		ret = append(ret, dst)
	}
	return ret, nil
}

func ConnectToDB(dbname string) {
	var err error
	dbConn, err = sql.Open("postgres", "sslmode=disable dbname="+dbname)
	if err != nil {
		log.Fatalf("Error connecting to postgres: %v\n", err)
	}
}

func GetDBConn() *sql.DB {
	return dbConn
}

func GetSegmentConfigs(conn *sql.DB) ([][]string, error) {
	var version string
	err := dbConn.QueryRow("select substring(version(),39,1)").Scan(&version)
	if err != nil {
		return nil, err
	}
	var query string
	if version == "6" {
		query = "select dbid, content, role, preferred_role, mode, status, port, hostname, address, datadir as fselocation  from pg_catalog.gp_segment_configuration  order by content asc, role desc;"
	} else {
		query = "select dbid, content, role, preferred_role, mode, status, port, hostname, address, replication_port, fselocation  from pg_catalog.gp_segment_configuration c join pg_catalog.pg_filespace_entry fse on c.dbid = fse.fsedbid join pg_catalog.pg_filespace fs on fs.oid = fse.fsefsoid where fs.fsname = 'pg_system' order by content asc, role desc;"
	}

	rows, err := dbConn.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return GetResultsArray(rows)
}

func DumpQueryResultsToCSV(conn *sql.DB, query string, cw *csv.Writer) error {
	rows, qErr := conn.Query(query)
	if qErr != nil {
		return qErr
	}
	defer rows.Close()
	colNames, _ := rows.Columns()
	cw.Write(colNames)
	cw.Flush()
	readCols := make([]interface{}, len(colNames))
	writeCols := make([]sql.NullString, len(colNames))
	for i, _ := range readCols {
		readCols[i] = &writeCols[i]
	}
	out := make([]string, len(writeCols))
	for rows.Next() {
		if rErr := rows.Scan(readCols...); rErr != nil {
			return rErr
		}
		for i := range writeCols {
			out[i] = writeCols[i].String
		}
		cw.Write(out)
		cw.Flush()
	}
	cw.Flush()
	return nil
}

/*
The pq library does not support Kerberos authentication at this time.

pq: unexpected error: "setting PGKRBSRVNAME not supported"

Until and unless Kerberos is supported by pq, we will disable this
authentication method.
*/
func DisableKerberos() error {
	if ksetting, set := os.LookupEnv("PGKRBSRVNAME"); set {
		log.Println("PGKRBSRVNAME environment variable is set to %s. Kerberos is not supported. Disabling.", ksetting)
		err := os.Unsetenv("PGKRBSRVNAME")
		if err != nil {
			return err
		}
	}
	return nil
}

/*
exported function to get major version of gpdb
*/
func GetMajorVersion(db *sql.DB) (int, error) {
	var err error

	// Connect to template1 if db value is nil
	if db == nil {
		db, err = GPDBConnect("template1")
		if err != nil {
			return -1, err
		}
		defer db.Close()
	}

	var versionstring string
	err = db.QueryRow("select version();").Scan(&versionstring)
	if err != nil {
		return -1, err
	}

	r, _ := regexp.Compile(`Greenplum Database (\d+)\.\d+\.\d+`)
	regexresult := r.FindStringSubmatch(versionstring)
	if len(regexresult) == 2 {
		return strconv.Atoi(regexresult[1])
	} else {
		return -1, errors.New("Could not find version string")
	}
}
