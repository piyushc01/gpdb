#!/usr/bin/env python3

from gppylib.recoveryinfo import RecoveryErrorType
from gppylib.commands.pg import PgBaseBackup, PgRewind
from recovery_base import RecoveryBase, set_recovery_cmd_results
from gppylib.commands.base import Command
from gppylib.commands.gp import SegmentStart
from gppylib.gparray import Segment
from gppylib.commands.gp import ModifyConfSetting
from gppylib.db import dbconn


class FullRecovery(Command):
    def __init__(self, name, recovery_info, forceoverwrite, logger, era):
        self.name = name
        self.recovery_info = recovery_info
        self.replicationSlotName = 'internal_wal_replication_slot'
        self.forceoverwrite = forceoverwrite
        self.era = era
        # FIXME test for this cmdstr. also what should this cmdstr be ?
        cmdStr = ''
        #cmdstr = 'TODO? : {} {}'.format(str(recovery_info), self.verbose)
        Command.__init__(self, self.name, cmdStr)
        #FIXME this logger has to come after the init and is duplicated in all the 4 classes
        self.logger = logger
        self.error_type = RecoveryErrorType.DEFAULT_ERROR

    @set_recovery_cmd_results
    def run(self):
        self.error_type = RecoveryErrorType.BASEBACKUP_ERROR
        cmd = PgBaseBackup(self.recovery_info.target_datadir,
                           self.recovery_info.source_hostname,
                           str(self.recovery_info.source_port),
                           create_slot=False,
                           replication_slot_name=self.replicationSlotName,
                           forceoverwrite=self.forceoverwrite,
                           target_gp_dbid=self.recovery_info.target_segment_dbid,
                           progress_file=self.recovery_info.progress_file)
        cmd.set_cmd_id = self.recovery_info.target_segment_dbid
        self.logger.info("Running pg_basebackup with progress output temporarily in %s" % self.recovery_info.progress_file)
        try:
            cmd.run(validateAfter=True)
        except Exception as e: #TODO should this be ExecutionError?
            self.logger.info("Running pg_basebackup failed: {}".format(str(e)))

            #  If the cluster never has mirrors, cmd will fail
            #  quickly because the internal slot doesn't exist.
            #  Re-run with `create_slot`.
            #  GPDB_12_MERGE_FIXME could we check it before? or let
            #  pg_basebackup create slot if not exists.
            cmd = PgBaseBackup(self.recovery_info.target_datadir,
                               self.recovery_info.source_hostname,
                               str(self.recovery_info.source_port),
                               create_slot=True,
                               replication_slot_name=self.replicationSlotName,
                               forceoverwrite=True,
                               target_gp_dbid=self.recovery_info.target_segment_dbid,
                               progress_file=self.recovery_info.progress_file)
            self.logger.info("Re-running pg_basebackup, creating the slot this time")
            cmd.run(validateAfter=True)

        self.error_type = RecoveryErrorType.DEFAULT_ERROR
        self.logger.info("Successfully ran pg_basebackup for dbid: {}".format(
            self.recovery_info.target_segment_dbid))

        cmd_runtime = cmd.get_cmd_runtime()
        #Push this data to segment DB:
        dburl = dbconn.DbURL(hostname=self.recovery_info.source_hostname,
                             port=self.recovery_info.source_port, dbname='template1')
        conn = dbconn.connect(dburl, utility=True)
        dbconn.execSQL(conn, "CREATE TABLE IF NOT EXISTS recovery_history_table("
                             "ID  SERIAL PRIMARY KEY, SOURCE_DBID INTEGER, DEST_DBID INTEGER, RECOVERY_IS_FULL INTEGER ,"
                             "RECOVERY_TIME INTEGER, NETWORK_SPPED_AT_START INTEGER , DISK_READ_SPEED INTEGER, "
                             "RECOVERY_SIZE INTEGER );")

        postgres_insert_query = "INSERT INTO recovery_history_table(SOURCE_DBID, DEST_DBID, RECOVERY_IS_FULL, RECOVERY_TIME," \
                                " METWORK_SPEED_AT_START, DISK_READ_SPEED, RECOVERY_SIZE) VALUES (%d, %d, 1, %d, 0, 0, 0)" \
                                %(self.recovery_info.target_segment_dbid, self.recovery_info.target_segment_dbid, cmd_runtime)

        dbconn.execSQL(conn, postgres_insert_query)
        conn.close()

        # Updating port number on conf after recovery
        self.error_type = RecoveryErrorType.UPDATE_ERROR
        update_port_in_conf(self.recovery_info, self.logger)

        self.error_type = RecoveryErrorType.START_ERROR
        start_segment(self.recovery_info, self.logger, self.era)


class IncrementalRecovery(Command):
    def __init__(self, name, recovery_info, logger, era):
        self.name = name
        self.recovery_info = recovery_info
        self.era = era
        cmdStr = ''
        Command.__init__(self, self.name, cmdStr)
        self.logger = logger
        self.error_type = RecoveryErrorType.DEFAULT_ERROR

    def get_network_speed(self, hostname):
        cmd_str = "dd if=/dev/zero of=test bs=1024k count=200 2>/dev/null;rsync -av test %s:test|tail -2|head -1" % hostname
        cmd = Command(name='Get-Network-Speed', cmdStr=cmd_str, ctxt=REMOTE, remoteHost=hostname)
        cmd.run(validateAfter=True)
        str_output = cmd.get_results().stdout
        # sent 58060 bytes  received 101430 bytes  106326.67 bytes/sec

    @set_recovery_cmd_results
    def run(self):
        self.logger.info("Running pg_rewind with progress output temporarily in %s" % self.recovery_info.progress_file)

        self.error_type = RecoveryErrorType.REWIND_ERROR
        cmd = PgRewind('rewind dbid: {}'.format(self.recovery_info.target_segment_dbid),
                       self.recovery_info.target_datadir, self.recovery_info.source_hostname,
                       self.recovery_info.source_port, self.recovery_info.progress_file)
        cmd.run(validateAfter=True)
        self.logger.info("Successfully ran pg_rewind for dbid: {}".format(self.recovery_info.target_segment_dbid))
        cmd_runtime = cmd.get_cmd_runtime()
        # Push this data to segment DB:
        dburl = dbconn.DbURL(hostname=self.recovery_info.source_hostname, port=self.recovery_info.source_port, dbname='template1')
        conn = dbconn.connect(dburl, utility=True)
        dbconn.execSQL(conn, "CREATE SCHEMA IF NOT EXISTS gpdb_test;")

        dbconn.execSQL(conn, "CREATE TABLE IF NOT EXISTS gpdb_test.recovery_history_table("
                             "ID  SERIAL PRIMARY KEY, SOURCE_DBID INTEGER, DEST_DBID INTEGER, RECOVERY_IS_FULL INTEGER ,"
                             "RECOVERY_TIME INTEGER, NETWORK_SPEED_AT_START INTEGER , DISK_READ_SPEED INTEGER, "
                             "RECOVERY_SIZE INTEGER );")
        postgres_insert_query = "INSERT INTO gpdb_test.recovery_history_table(SOURCE_DBID, DEST_DBID, RECOVERY_IS_FULL, RECOVERY_TIME," \
                                " NETWORK_SPEED_AT_START, DISK_READ_SPEED, RECOVERY_SIZE) VALUES (%d, %d, 0, %d, 0, 0, 0)" \
                                % (self.recovery_info.target_segment_dbid, self.recovery_info.target_segment_dbid,
                                   cmd_runtime)
        dbconn.execSQL(conn, postgres_insert_query)
        conn.close()
        # Updating port number on conf after recovery
        self.error_type = RecoveryErrorType.UPDATE_ERROR
        update_port_in_conf(self.recovery_info, self.logger)

        self.error_type = RecoveryErrorType.START_ERROR
        start_segment(self.recovery_info, self.logger, self.era)


def start_segment(recovery_info, logger, era):
    seg = Segment(None, None, None, None, None, None, None, None,
                  recovery_info.target_port, recovery_info.target_datadir)
    cmd = SegmentStart(
        name="Starting new segment with dbid %s:" % (str(recovery_info.target_segment_dbid))
        , gpdb=seg
        , numContentsInCluster=0
        , era=era
        , mirrormode="mirror"
        , utilityMode=True)
    logger.info(str(cmd))
    cmd.run(validateAfter=True)


def update_port_in_conf(recovery_info, logger):
    logger.info("Updating %s/postgresql.conf" % recovery_info.target_datadir)
    modifyConfCmd = ModifyConfSetting('Updating %s/postgresql.conf' % recovery_info.target_datadir,
                                      "{}/{}".format(recovery_info.target_datadir, 'postgresql.conf'),
                                      'port', recovery_info.target_port, optType='number')
    modifyConfCmd.run(validateAfter=True)


#FIXME we may not need this class
class SegRecovery(object):
    def __init__(self):
        pass

    def main(self):
        recovery_base = RecoveryBase(__file__)
        recovery_base.main(self.get_recovery_cmds(recovery_base.seg_recovery_info_list, recovery_base.options.forceoverwrite,
                                                  recovery_base.logger, recovery_base.options.era))

    def get_recovery_cmds(self, seg_recovery_info_list, forceoverwrite, logger, era):
        cmd_list = []
        for seg_recovery_info in seg_recovery_info_list:
            if seg_recovery_info.is_full_recovery:
                cmd = FullRecovery(name='Run pg_basebackup',
                                   recovery_info=seg_recovery_info,
                                   forceoverwrite=forceoverwrite,
                                   logger=logger,
                                   era=era)
            else:
                cmd = IncrementalRecovery(name='Run pg_rewind',
                                          recovery_info=seg_recovery_info,
                                          logger=logger,
                                          era=era)
            cmd_list.append(cmd)
        return cmd_list


if __name__ == '__main__':
    SegRecovery().main()
