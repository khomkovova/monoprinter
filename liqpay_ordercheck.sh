#!/bin/bash
PUBLIC_KEY='sandbox_i91396305288'
PRIVATE_KEY='sandbox_koXlZthN7pe2WULIegJrhusEbsBG0U3OZnwlkBki'
API_URL='https://www.liqpay.ua/api/request'
JSON="{ 
\"action\" : \"status\",
\"version\" : 3,
\"public_key\" : \"${PUBLIC_KEY}\", 
\"order_id\" : \"LuQ74tdAC1F0oRiRkjTogoXC4ESdDklf0vzjoXL8C94N76Kozr7NFQyrZyer2BidgtpH5ejO0/9bK34zSvvjQOtcuQBCvT1bGiCTdjOlomUn67O1Qh2FbvYwOiifKS8ivEpC\"
}"
# DATA is base64_encode result from JSON string
DATA=$(echo -n ${JSON} | base64)
# SIGNATURE is base64 encode result from sha1 binary hash from concatenate string ${PRIVATE_KEY}${DATA}${PRIVATE_KEY}
SIGNATURE=$(echo -n "${PRIVATE_KEY}${DATA}${PRIVATE_KEY}" | openssl dgst -binary -sha1 | base64)
# REQ is json response from liqpay
REQ=$(curl --silent -XPOST ${API_URL} --data-urlencode data="${DATA}" --data-urlencode signature="${SIGNATURE}")
echo "Result: ${REQ}"