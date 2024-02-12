#!/bin/bash

SQL_COMMAND="SELECT lr.id AS latest_release_id, 
       lr.name AS release_name, 
       a.name AS artist_name
FROM latest_releases lr
JOIN artists a ON lr.artist_id = a.id;"

docker exec -it yamb-db-dev psql -U yamb -d yamb -c "$SQL_COMMAND"
