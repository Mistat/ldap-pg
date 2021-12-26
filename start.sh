#!/bin/sh
./bin/ldap-pg \
  -acl ou=Users,dc=shidax:R:sn,userPassword \
  -h $DB_HOST \
  -u $DB_USERNAME \
  -w $DB_PASSWORD \
  -d $DB_NAME \
  -p $DB_PORT \
  -s public \
  -b 0.0.0.0:$PORT \
  -root-dn $USERNAME \
  -root-pw $PASSWORD \
  -log-level $LOGLEVEL