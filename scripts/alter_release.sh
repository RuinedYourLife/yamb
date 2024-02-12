#!/bin/bash

if [ -z "$1" ]; then
    echo "Please provide the release id as an argument."
    exit 1
fi

SQL_COMMAND="update latest_releases set release_date = '2023-01-01' where id = $1;"

docker exec -it yamb-db-dev psql -U yamb -d yamb -c "$SQL_COMMAND"
