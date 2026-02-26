#!/usr/bin/bash

if [[ $# -ne 1 ]]; then
	printf "Usage:\n  ./start_containers.sh <num_containers>"
	exit
fi

i=0
s="services:\n"

while [[ $i -lt $1 ]]; do
    s+="  service${i}:\n"
    s+="    build:\n"
    s+="      context: ./app\n"
    s+="    networks:\n"
    s+="      net:\n"
    s+="        ipv4_address: 10.10.${i}.1\n"
    i=$((i + 1))
done

s+="\nnetworks:\n"
s+="  net:\n"
s+="    ipam:\n"
s+="      driver: default\n"
s+="      config:\n"
s+="        - subnet: 10.10.0.0/16\n"
s+="          ip_range: 10.10.0.0/16\n"
s+="          gateway: 10.10.0.254\n"

printf "$s" > ./docker-compose.yml

docker compose build
docker compose up
