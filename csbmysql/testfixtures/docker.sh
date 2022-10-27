#!/usr/bin/env bash

wd="$HOME/workspace/csb/terraform-provider-csbmysql/csbmysql/testfixtures/"
docker volume create mysql_config
for d in certs keys; do
  docker run -v "${wd}:/fixture" --mount source=mysql_config,destination=/mnt mysql rm -rf "/mnt/$d"
  docker run -v "${wd}:/fixture" --mount source=mysql_config,destination=/mnt mysql cp -r "/fixture/ssl_mysql/$d" /mnt
done
docker run -v "${wd}:/fixture" --mount source=mysql_config,destination=/mnt mysql rm "/mnt/my.cnf"
docker run -v "${wd}:/fixture" --mount source=mysql_config,destination=/mnt mysql cp "/fixture/my.cnf" "/mnt"
docker run -v "${wd}:/fixture" --mount source=mysql_config,destination=/mnt mysql chown mysql "/mnt/keys/server.key"
docker run -v "${wd}:/fixture" --mount source=mysql_config,destination=/mnt mysql chmod 0600 "/mnt/keys/server.key"

docker run --name=mysql \
  --publish=3306:3306 \
  --env MYSQL_ROOT_PASSWORD="password" \
  --mount source=mysql_config,destination=/etc/mysql/conf.d \
  --health-cmd 'mysqladmin -h 127.0.0.1 -P 3306 -u root -ppassword' \
  mysql:5.7
