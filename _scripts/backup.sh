#!/bin/bash
#
# Backup and upload to S3
#
# Usage:
#   $ backup.sh <backup directory>
#
# If FORCE_BACKUP is set to 1, the backup will be done even if the
# backup directory hash is the same as the previous backup.

HASHFILE=/tmp/backuphash
#S3_BUCKET=s3://<your-s3-bucket>
#BACKUP_FILE=/tmp/backupfilename.tar.gz

# calc md5 of a directory
md5() {
    find "$1" -type f -exec md5sum {} \; | sort -k 2 | md5sum | cut -d ' ' -f 1
}

# store last md5
save() {
    echo $1 > $2
}

# get last md5
get() {
    cat $1
}

tar() {
    tarfile=$1
    target=$2
    /bin/tar -zcf $tarfile $target
}

backup() {
    # check if need backup
    local need_backup=0
    local target=$1
    local backup_file=$2
    local s3_bucket=$3
    local hash=$(md5 $target)
    if [ -f $HASHFILE ]; then
        if [ $(get $HASHFILE) != $hash ]; then
            need_backup=1
        else 
            need_backup=0
            echo "checksum not changed"
        fi
    else 
        need_backup=1
    fi

    # tar and upload
    if [[ $need_backup -eq 1 || $FORCE_BACKUP -eq 1 ]]; then
        echo "start backup..."
        tar $backup_file $target
        echo "uploading..."
        if [ -f $backup_file ]; then
            local basename=$(basename $backup_file)
            aws s3 cp $backup_file $s3_bucket/$basename
            save $hash $HASHFILE
            # remove tar file
            rm $backup_file
            echo "backup success"
        else
            echo "backup failed."
        fi
    fi
}

backup $1 $BACKUP_FILE $S3_BUCKET
exit 0
