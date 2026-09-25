-- Wire the local onion MySQL into MindsDB as a queryable datasource.
-- Run via: make local-env-mindsdb-init
CREATE DATABASE IF NOT EXISTS onion_db
WITH ENGINE = 'mysql',
PARAMETERS = {
  "host": "db",
  "port": 3306,
  "user": "root",
  "password": "onion_root_password",
  "database": "onion_dev"
};
