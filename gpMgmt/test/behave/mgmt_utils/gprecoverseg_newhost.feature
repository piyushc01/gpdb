@gprecoverseg_newhost
Feature: gprecoverseg tests involving migrating to a new host

########################### @concourse_cluster tests ###########################
# The @concourse_cluster tag denotes the scenario that requires a remote cluster

    # TODO: There is a false dependency on PGDATABASE=gptest in our behave tests, so we create it here.
    @concourse_cluster
    Scenario Outline: "gprecoverseg -p newhosts" successfully recovers for <test_case>
      Given the database is running
      And all the segments are running
      And the segments are synchronized
      And database "gptest" exists
      And the user runs gpconfig sets guc "wal_sender_timeout" with "15s"
      And the user runs "gpstop -air"
      And the cluster configuration is saved for "before"
      And segment hosts <down> are disconnected from the cluster and from the spare segment hosts <spare>
      And the cluster configuration has no segments where <down_sql>
      When the user runs <gprecoverseg_cmd>
      Then gprecoverseg should return a return code of 0
      And pg_hba file "/data/gpdata/mirror/gpseg0/pg_hba.conf" on host "<acting_primary>" contains entries for "<used>"
      And the cluster configuration is saved for "<test_case>"
      And the "before" and "<test_case>" cluster configuration matches with the expected for gprecoverseg newhost
      And the mirrors replicate and fail over and back correctly
      And the cluster is rebalanced
      And the original cluster state is recreated for "<test_case>"
      And the cluster configuration is saved for "after_recreation"
      And the "before" and "after_recreation" cluster configuration matches with the expected for gprecoverseg newhost
      Examples:
      | test_case      |  down        | spare | unused | used | acting_primary | gprecoverseg_cmd                              | down_sql                                              |
      | one_host_down  |  "sdw1"      | "sdw5" | "sdw6"   | sdw5 | sdw2           | "gprecoverseg -a -p sdw5 --hba-hostnames"   | "hostname='sdw1' and status='u'"                      |
      | two_hosts_down |  "sdw1,sdw3" | "sdw5,sdw6" | none   | sdw5 | sdw2           | "gprecoverseg -a -p sdw5,sdw6 --hba-hostnames" | "(hostname='sdw1' or hostname='sdw3') and status='u'" |

    @concourse_cluster
    Scenario: failed host is not in reach gprecoverseg recovery works well with all instances recovered
       Given  the database is running
       And all the segments are running
       And the segments are synchronized
       And database "gptest" exists
       And the cluster configuration is saved for "before"
       And segment hosts "sdw1" are disconnected from the cluster and from the spare segment hosts "sdw5,sdw6"
       And the user runs psql with "-c 'SELECT gp_request_fts_probe_scan()'" against database "postgres"
       And the cluster configuration has no segments where "hostname='sdw1' and status='u'"
       When the user runs command "echo 'sdw1|20000|/data/gpdata/primary/gpseg0 sdw5|20000|/data/gpdata/primary/gpseg0' > /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs command "echo 'sdw1|21001|/data/gpdata/mirror/gpseg7 sdw5|21001|/data/gpdata/mirror/gpseg7' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs command "echo 'sdw1|21000|/data/gpdata/mirror/gpseg6 sdw5|21000|/data/gpdata/mirror/gpseg6' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs command "echo 'sdw1|20001|/data/gpdata/primary/gpseg1 sdw5|20001|/data/gpdata/primary/gpseg1' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs "gprecoverseg -i /tmp/test-gprecoverseg01-scheraio-config-file -av"
       Then gprecoverseg should return a return code of 0
       #Then segment hosts "sdw1" are reconnected to the cluster and to the spare segment hosts "sdw6"
       Then the original cluster state is recreated for "one_host_down_-i"
       #And database "gptest" exists
       And the cluster configuration is saved for "after_recreation"
       And the "before" and "after_recreation" cluster configuration matches with the expected for gprecoverseg newhost

    @concourse_cluster
    Scenario: failed host is not in reach gprecoverseg recovery works well with partial recovery
       Given  the database is running
       And all the segments are running
       And the segments are synchronized
       And database "gptest" exists
       And the cluster configuration is saved for "before"
       And segment hosts "sdw1" are disconnected from the cluster and from the spare segment hosts "sdw5,sdw6"
       And the user runs psql with "-c 'SELECT gp_request_fts_probe_scan()'" against database "postgres"
       And the cluster configuration has no segments where "hostname='sdw1' and status='u'"
       When the user runs command "echo 'sdw1|20000|/data/gpdata/primary/gpseg0 sdw5|20000|/data/gpdata/primary/gpseg0' > /tmp/test-gprecoverseg01-scheraio-config-file"
       #Then the user runs command "echo 'sdw2|20000|/data/gpdata/primary/gpseg2 sdw5|20000|/data/gpdata/primary/gpseg2' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs command "echo 'sdw1|21000|/data/gpdata/mirror/gpseg6 sdw5|21000|/data/gpdata/mirror/gpseg6' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       #Then the user runs command "echo 'sdw2|21001|/data/gpdata/mirror/gpseg1 sdw5|21001|/data/gpdata/mirror/gpseg1' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs "gprecoverseg -i /tmp/test-gprecoverseg01-scheraio-config-file -av"
       Then gprecoverseg should return a return code of 0
       Then the original cluster state is recreated for "one_host_down_-i"
       #Then segment hosts "sdw1" are reconnected to the cluster and to the spare segment hosts "sdw6"
       #And database "gptest" exists
       And the cluster configuration is saved for "after_recreation"
       And the "before" and "after_recreation" cluster configuration matches with the expected for gprecoverseg newhost

    @concourse_cluster
    Scenario: failed host is not in reach gprecoverseg recovery works well only primaries are recovered
       Given  the database is running
       And all the segments are running
       And the segments are synchronized
       And database "gptest" exists
       And the cluster configuration is saved for "before"
       And segment hosts "sdw1" are disconnected from the cluster and from the spare segment hosts "sdw5,sdw6"
       And the user runs psql with "-c 'SELECT gp_request_fts_probe_scan()'" against database "postgres"
       And the cluster configuration has no segments where "hostname='sdw1' and status='u'"
       When the user runs command "echo 'sdw1|20000|/data/gpdata/primary/gpseg0 sdw5|20000|/data/gpdata/primary/gpseg0' > /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs command "echo 'sdw1|21001|/data/gpdata/mirror/gpseg7 sdw5|21001|/data/gpdata/mirror/gpseg7' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       #Then the user runs command "echo 'sdw2|21000|/data/gpdata/mirror/gpseg0 sdw5|21000|/data/gpdata/mirror/gpseg0' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       #Then the user runs command "echo 'sdw2|21001|/data/gpdata/mirror/gpseg1 sdw5|21001|/data/gpdata/mirror/gpseg1' >> /tmp/test-gprecoverseg01-scheraio-config-file"
       Then the user runs "gprecoverseg -i /tmp/test-gprecoverseg01-scheraio-config-file -av"
       Then gprecoverseg should return a return code of 0
       Then the original cluster state is recreated for "one_host_down_-i"
       #Then segment hosts "sdw1" are reconnected to the cluster and to the spare segment hosts "sdw6"
       #Then the original cluster state is recreated for "one_host_down"
       #And database "gptest" exists
       And the cluster configuration is saved for "after_recreation"
       And the "before" and "after_recreation" cluster configuration matches with the expected for gprecoverseg newhost
