---
title: Upgrading from Greenplum 6 to Greenplum 7 
---

This topic walks you through upgrading from Greenplum 6 to Greenplum 7. The upgrade procedure varies, depending on whether you are upgrading the software while remaining on the same hardware or upgrading the software while moving to new hardware.

>**WARNING**
>There are a substantial number of changes between Greenplum 6 and Greenplum 7 that could potentially affect your existing application when you move to Greenplum 7. Before going any further, familiarize yourself with all of these changes [Important Changes between Greenplum 6 and Greenplum 7](./changes-6-7-landing-page.html).

After reading through [Preparing to Upgrade](#preparing), choose one of these two paths:

[Upgrading to Greenplum 7 on the Same Hardware](#same_hardware)<br>
[Upgrading to Greenplum 7 While Moving to New Hardware](#new_hardware)

>**Note**
>You cannot upgrade to VMware Greenplum 7.0.0 using the `gpupgrade` utility.

## <a id="preparing"></a>Preparing to Upgrade

To prepare for upgrading from Greenplum 6 to Greenplum 7:

- Read carefully through [Important Changes between Greenplum 6 and Greenplum 7](./changes-6-7-landing-page.html) to identify changes you may need to make in your application before upgrading to Greenplum 7.

- Before using `gpbackup`/`gprestore` to move your data, read carefully through [Backup and Restore Caveats](#br-caveats) to avert problems that can arise.  

>**Note**
You must use `gpbackup`/`gprestore` if upgrading on the same hardware. You may choose between `gpbackup`/`gprestore` and `gpcopy` when upgrading while moving to new hardware.

## <a id="same_hardware"></a>Upgrading to Greenplum 7 On the Same Hardware

Follow the steps below to upgrade to Greenplum 7 while remaining on the same hardware. You will move your data using the `gpbackup/gprestore` utilities. There are a number of caveats with respect to backing up and restoring your data. Be sure to review [Backup and Restore Caveats](#backup-and-restore-caveats) before starting the upgrade.

To upgrade while staying on the same hardware:

1. If not already installed, install the latest release of the Greenplum Backup and Restore utilities, available to download from [VMware Tanzu Network](https://network.pivotal.io/products/greenplum-backup-restore) or [github](https://github.com/greenplum-db/gpbackup/releases).

2. Run the `gpbackup` utility to back up your data to an external data source. For more infomation on `gpbackup`, see the [VMware Greenplum Backup and Restore guide](https://docs.vmware.com/en/VMware-Greenplum-Backup-and-Restore/1.29/greenplum-backup-and-restore/backup-restore.html).

3. Delete the Greenplum 6 cluster by issuing the [`gpdeletesystem` command](../utility_guide/ref/gpdeletesystem.html).

4. Initialize a Greenplum 7 cluster by issuing the [`gpinitsystem` command](../utility_guide/ref/gpinitsystem.html).

5. Install any external modules used in your Greenplum 6 system in the Greenplum 7 system before you restore the backup, for example MADlib or PostGIS. If versions of the external modules are not compatible, you may need to exclude tables that reference them when restoring the Greenplum 6 backup to Greenplum 7.

6. Run the `gprestore` utility to restore your data to the Greenplum 7 cluster from the external data source. For more infomation on `gprestore`, see the [VMware Greenplum Backup and Restore guide](https://docs.vmware.com/en/VMware-Greenplum-Backup-and-Restore/1.29/greenplum-backup-and-restore/backup-restore.html).

>**Note**
>- When restoring language-based user-defined functions, the shared object file must be in the location specified in the `CREATE FUNCTION` SQL command and must have been recompiled on the Greenplum 7 system. This applies to user-defined functions, user-defined types, and any other objects that use custom functions, such as aggregates created with the `CREATE AGGREGATE` command.

## <a id="new_hardware"></a>Upgrading to Greenplum 7 While Moving to New Hardware

Follow the steps below to upgrade to Greenplum 7 while moving to new hardware. 

If you are upgrading from Greenplum 6 to Greenplum 7 and changing hardware you may move your data in one of two ways:

- By using the `gpbackup/gprestore` utilities
- By using the `gpcopy` utility

To upgrade while moving to new hardware using `gpbackup/gprestore`:

1. If not already installed, install the latest release of the Greenplum Backup and Restore utilities, available to download from [VMware Tanzu Network](https://network.pivotal.io/products/greenplum-backup-restore) or [github](https://github.com/greenplum-db/gpbackup/releases).

2. Run the `gpbackup` utility to back up the date from your Greenplum 6 cluster to an external data source. For more infomation on `gpbackup`, see the [VMware Greenplum Backup and Restore guide](https://docs.vmware.com/en/VMware-Greenplum-Backup-and-Restore/1.29/greenplum-backup-and-restore/backup-restore.html).

3. Initalize a Greenplum 7 cluster on the new hardware, by issuing the [`gpinitsystem` command](../utility_guide/ref/gpinitsystem.html).

4. Install any external modules used in your Greenplum 6 system in the Greenplum 7 system before you restore the backup, for example MADlib or PostGIS. If versions of the external modules are not compatible, you may need to exclude tables that reference them when restoring the Greenplum 6 backup to Greenplum 7.

5. Run the `gprestore` utility to restore your data to the Greenplum 7 cluster from the external data source. For more infomation on `gprestore`, see the [VMware Greenplum Backup and Restore guide](https://docs.vmware.com/en/VMware-Greenplum-Backup-and-Restore/1.29/greenplum-backup-and-restore/backup-restore.html).

>**Note**
>- When restoring language-based user-defined functions, the shared object file must be in the location specified in the `CREATE FUNCTION` SQL command and must have been recompiled on the Greenplum 7 system. This applies to user-defined functions, user-defined types, and any other objects that use custom functions, such as aggregates created with the `CREATE AGGREGATE` command.

To upgrade while moving to new hardware using `gpcopy`:

1. Review the information in [Migrating Data with gpcopy](https://docs.vmware.com/en/VMware-Greenplum-Data-Copy-Utility/2.6/greenplum-copy/gpcopy-migrate.html).

2. Initalize a Greenplum 7 cluster on the new hardware, by issuing the [`gpinitsystem` command](../utility_guide/ref/gpinitsystem.html).

3. Verify network connectivity between your Greenplum 6 and your Greenplum 7 cluster. 

4. Run the `gpcopy` utility to migrate your data to the Greenplum 7 cluster. 

## <a id="completing"></a>Completing the Upgrade

Migrate any tables you skipped during the restore using other methods, for example using the `COPY TO` command to create an external file and then loading the data from the external file into Greenplum 6 with the `COPY FROM` command.

Recreate any objects you dropped in the Greenplum 6 database to enable migration, such as external tables, indexes, user-defined functions, or user-defined aggregates.

After migrating data you may need to modify SQL scripts, administration scripts, and user-defined functions as necessary to account for changes in Greenplum Database version 7. Review the [VMware Greenplum 7.0.0 Release Notes](https://docs.vmware.com/en/VMware-Greenplum/7/greenplum-database/relnotes-release-notes.html#release-7.0.0) and [Important Changes between Greenplum 6 and Greenplum 7](./changes-6-7-landing-page.html) for features and changes that may necessitate post-migration tasks.

## <a id="br-caveats"></a>Backup and Restore Caveats

There are a number of caveats with respect to backing up and restoring your data as part of upgrading from Greenplum 6 to Greenplum 7.

- Before you do an actual backup, use `gpbackup` to create a `--metadata-only` backup from the source Greenplum database and restore it to the Greenplum 7 system. Review the `gprestore` log file for error messages and correct any remaining problems in the source Greenplum database.

- If you intend to install VMware Greenplum 7 on the same hardware as your 6 system, you will need enough disk space to accommodate over five times the original data set (two full copies of the primary and mirror data sets, plus the original backup data in ASCII format) in order to migrate data with `gpbackup` and `gprestore`. Keep in mind that the ASCII backup data will require more disk space than the original data, which may be stored in compressed binary format. Offline backup solutions such as Dell EMC Data Domain can reduce the required disk space on each host.

- When restoring language-based user-defined functions, the shared object file must be in the location specified in the `CREATE FUNCTION` SQL command and must have been recompiled on the Greenplum 7 system. This applies to user-defined functions, user-defined types, and any other objects that use custom functions, such as aggregates created with the `CREATE AGGREGATE` command.

- `gpbackup` saves the distribution policy and distribution key for each table in the backup so that data can be restored to the same segment. If a table's distribution key in the Greenplum 6 database is incompatible with Greenplum 7, `gprestore` cannot restore the table to the correct segment in the Greenplum 6 database. This can happen if the distribution key in the older Greenplum release has columns with data types not allowed in Greenplum 7 distribution keys, or if the data representation for data types has changed or is insufficient for Greenplum 7 to generate the same hash value for a distribution key. You should correct these kinds of problems by altering distribution keys in the tables before you back up the Greenplum database.

**Parent topic:** [Installing and Upgrading Greenplum](install_guide.html)