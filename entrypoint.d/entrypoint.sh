#!/bin/bash
echo "URL: ${PUBKEYURL}"
curl -o ${PUBKEYPATH} ${PUBKEYURL} 2>&1

echo "Checking content of ${PUBKEYPATH}"
chmod 775 ${PUBKEYPATH}
cat ${PUBKEYPATH}

exec "$@"
