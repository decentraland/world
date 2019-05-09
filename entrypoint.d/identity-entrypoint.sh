#!/bin/bash

echo "Setting clients data..."
echo -e ${CONFIG_SERVICE_JSON} | sed 's/\\//g' > ${CLIENTS_DATA_PATH}
cat ${CLIENTS_DATA_PATH}
echo "...Done"

echo "Setting private key..."
echo -e ${PRIV_KEY} | sed 's/\\//g' > ${PRIV_KEY_PATH}
cat ${PRIV_KEY_PATH}

if [ -s ${PRIV_KEY_PATH} ]
then
     echo "...Done"
else
     echo "...Failed to write private key"
     exit 1
fi

exec ./build/identity $@
