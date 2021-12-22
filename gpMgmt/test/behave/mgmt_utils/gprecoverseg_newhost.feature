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
      And segment hosts <down> are reconnected to the cluster and to the spare segment hosts "<unused>"
      And the original cluster state is recreated after cleaning up <down> hosts
      And database "gptest" exists
      And the cluster configuration is saved for "after_recreation"
      And the "before" and "after_recreation" cluster configuration matches with the expected for gprecoverseg newhost
      Examples:
      | test_case      |  down        | spare       | unused | used | acting_primary | gprecoverseg_cmd                               | down_sql                                              |
      | one_host_down  |  "sdw1"      | "sdw5,sdw6" | sdw6   | sdw5 | sdw2           | "gprecoverseg -a -p sdw5 --hba-hostnames"      | "hostname='sdw1' and status='u'"                      |
      | two_hosts_down |  "sdw1,sdw3" | "sdw5,sdw6" | none   | sdw5 | sdw2           | "gprecoverseg -a -p sdw5,sdw6 --hba-hostnames" | "(hostname='sdw1' or hostname='sdw3') and status='u'" |

    @concourse_cluster
    Scenario: failed host is not in reach gprecoverseg recovery works well
        Given  the database is running
        And all the segments are running
        And the segments are synchronized
        And database "gptest" exists
        And segment hosts "sdw2" are disconnected from the cluster and from the spare segment hosts "sdw5,sdw6"
        #And the cluster configuration has no segments where "hostname='sdw2' and status='u'"
        When the user runs command "echo '%s|20000|/data/gpdata/primary/gpseg2 %s|20000|/data/gpdata/primary/gpseg2' > /tmp/test-gprecoverseg01-scheraio-config-file"
        Then the user runs "gprecoverseg -i /tmp/test-gprecoverseg01-scheraio-config-file -a"
        Then gprecoverseg should return a return code of 0
        Then segment hosts "sdw2" are reconnected to the cluster and to the spare segment hosts "sdw6"
        When the user runs "gprecoverseg -a -F"
        Then all the segments are running
        And the segments are synchronized
        And database "gptest" exists
