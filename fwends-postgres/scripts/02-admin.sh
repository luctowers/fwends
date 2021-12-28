#!/bin/sh
export IFS=","
for email in $ADMIN_EMAILS; do
  echo """
  INSERT INTO admins (email) VALUES ('$email');
  """ | psql --username $POSTGRES_USER --dbname $POSTGRES_DB
done
