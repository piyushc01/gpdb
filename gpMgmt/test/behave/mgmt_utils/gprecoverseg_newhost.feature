@gprecoverseg_newhost
Feature: gprecoverseg tests involving migrating to a new host

########################### @concourse_cluster tests ###########################
# The @concourse_cluster tag denotes the scenario that requires a remote cluster


    @concourse_cluster
    Scenario: failed host is not in reach gprecoverseg recovery works well
      Given  the database is running
      And all the segments are running
      And the segments are synchronized
      #And the cluster configuration is saved for "before"
      And segment hosts "sdw2" are disconnected from the cluster and from the spare segment hosts "sdw5,sdw6"
      And the user runs psql with "-c 'SELECT gp_request_fts_probe_scan()'" against database "postgres"
      #And the cluster configuration has no segments where "hostname='sdw2' and status='u'"
      When the user runs command "echo 'sdw2|20001|/data/gpdata/primary/gpseg3 sdw5|20001|/data/gpdata/primary/gpseg3' > /tmp/test-gprecoverseg01-scheraio-config-file"
      Then the user runs command "echo 'sdw2|20000|/data/gpdata/primary/gpseg2 sdw5|20000|/data/gpdata/primary/gpseg2' >> /tmp/test-gprecoverseg01-scheraio-config-file"
      #Then the user runs "gprecoverseg -i /tmp/test-gprecoverseg01-scheraio-config-file -av"
      #Then gprecoverseg should return a return code of 0
      Then segment hosts "sdw2" are reconnected to the cluster and to the spare segment hosts "sdw6"
      #Then the original cluster state is recreated after cleaning up "sdw2" hosts
      #And the cluster configuration is saved for "after_recreation"
      #And the "before" and "after_recreation" cluster configuration matches with the expected for gprecoverseg newhost
