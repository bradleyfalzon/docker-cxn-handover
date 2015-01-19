#!/bin/bash
set -o nounset

# TODO
# checkErrors

# Get the latest container id after running docker run
NEW=d8028265330d42c6c7ff3e56b252fcd5a8475daa3734f52e41ac60caea9ff443

CHAIN_PREFIX="SHRT"

# Return the port number that the container wan
function parse_listening_port {
	echo $1 | egrep -o '"[0-9]+/tcp"' | egrep -o '[0-9]+'
}

function parse_container_port {
	echo $1 | egrep -o '"[0-9]+"$' | egrep -o '[0-9]+'
}

function parse_container_ip {
	IP_JSON=$(docker inspect ${NEW} | ./JSON.sh -l | egrep '\[0,"NetworkSettings","IPAddress"\]')
	echo ${IP_JSON} | egrep -o '"[0-9.]+"$' | egrep -o '[0-9.]+'
}

function create_chain {
	RAND=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 4)
	iptables -t nat --new-chain "${CHAIN_PREFIX}_${RAND}"
	echo ${CHAIN_PREFIX}_${RAND}
}

# $1 = LISTEN, $2 = HOST_PORT
function add_nat_rules {
	iptables -t nat -I ${CHAIN} 1 ! --in-interface docker0 -p tcp --dport $LISTEN -j DNAT --to-destination ${CONTAINER_IP}:${LISTEN}
}

# Change the IFS to allow for loop to loop over new lines regardless of spaces
IFS=$(echo -en "\n\b")

# Fetch the new container's auto assigned IP
CONTAINER_IP=$(parse_container_ip)

# Create our new chain to store the rules for this new container
CHAIN=$(create_chain)
echo "New chain: ${CHAIN} for container IP: ${CONTAINER_IP}"

# Get a list of all the old IDs, these will be removed
OLD_IDS=$(docker ps --no-trunc=true --quiet=true | grep -v ${NEW})
echo "Old containers to be removed: $OLD_IDS"

# Find the new host port being used for this container
PORTS_JSON=$(docker inspect ${NEW} | ./JSON.sh -l | egrep '\[0,"NetworkSettings","Ports","[0-9]+/tcp",0,"HostPort"\]')

# Don't run if there's no ports exposed
if [ "$PORTS_JSON" == "" ]; then
	echo "No ports exposed for this container"
	exit 1
fi

# For each port that's exposed by the container, determine which port number
# the container is listening on, and what the random port number was auto
# assigned.
for i in $PORTS_JSON; do
	LISTEN=$(parse_listening_port $i)
	HOST=$(parse_container_port $i)
	echo "Listen: $LISTEN"

	# Add the port to the new chain
	add_nat_rules $LISTEN
done

# Set new chain to be the first entry point
echo "Adding new chain ${CHAIN} as the first entry"
#iptables -t nat -I DOCKER 1 -j ${CHAIN}
iptables -t nat -I DOCKER -j ${CHAIN}

echo "Deleting old chain(s)..."

OLD_CHAINS=$(iptables -t nat -L DOCKER | egrep -o "^${CHAIN_PREFIX}_[a-zA-Z0-9]+" | grep -v ${CHAIN})

if [ "$OLD_CHAINS" == "" ]; then
	echo "No old chains need to be deleted"
fi

OLD_CHAIN_NAMES=""
for i in $OLD_CHAINS; do
	echo "Deleting old DOCKER chain rule $i"
	iptables -t nat -D DOCKER -j $i
	OLD_CHAIN_NAMES="${i} ${OLD_CHAIN_NAMES}"
done

# Finally, delete the unused chains
OLD_CHAIN_NAMES_UNIQ=$(echo ${OLD_CHAIN_NAMES} | tr " " "\n" | sort | uniq)
for i in ${OLD_CHAIN_NAMES_UNIQ}; do
	echo "Deleting old chain $i"
	iptables -t nat -F ${i}
	iptables -t nat -X ${i}
done
