-- Greenplum specific access method tests, in addition to what's
-- covered by upstream create_am.sql tests

\set HIDE_TABLEAM off

create table create_am_gp_ao1 (a int, b int) using ao_row distributed by (a);
\d+ create_am_gp_ao1

create table create_am_gp_ao2 (a int, b int) using ao_row with (compresstype=zlib) distributed by (a);
\d+ create_am_gp_ao2

-- Should fail
create table create_am_gp_ao3 (a int, b int) using ao_row with (compresstype=rle_type) distributed by (a);
create table create_am_gp_ao3 (a int, b int) using heap with (compresstype=rle_type) distributed by (a);

create table create_am_gp_ao3 (a int, b int) using ao_column with (compresstype=rle_type) distributed by (a);
\d+ create_am_gp_ao3

-- Should fail because encoding clause is not supported by the tableam
create table create_am_gp_ao4(a int, b int encoding (compresstype=zlib)) using ao_row distributed by (a);

set gp_default_storage_options='compresstype=rle_type';

create table create_am_gp_heap(a int, b int) using heap distributed by (a);
-- should not have compresstype parameter
\d+ create_am_gp_heap

-- Should fail because the default compresstype configured above is
-- not supported by this tableam
create table create_am_gp_ao5(a int, b int) using ao_row distributed by (a);
create table create_am_gp_ao6(a int, b int) using ao_row with (compresstype=zlib) distributed by (a);
\d+ create_am_gp_ao6

create table create_am_gp_ao7(a int, b int encoding (compresstype=zlib)) using ao_column distributed by (a);
\d+ create_am_gp_ao7
