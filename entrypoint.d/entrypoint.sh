#!/bin/bash
FILE="config/enabledClients.json"

echo "Overwriting the file ${FILE}"
echo -e ${CONFIG_SERVICE_JSON} | sed 's/\\//g' > ${FILE}
cat ${FILE}

echo "Setting private key..."
echo -e ${PRIV_KEY} | sed 's/\\//g' > ${PRIV_KEY_PATH}

if [ -s ${PRIV_KEY_PATH} ]
then
     echo "...Done"
else
     echo "...Failed to write private key"
     exit 1
fi

exec "$@"
