#!/bin/sh
set -eu

curl --fail-with-body -X POST http://localhost:8080/index \
  -H 'Content-Type: application/json' \
  -d '[{"id":"checkout-1042","order_id":"1042","kind":"checkout","customer":"c-7","text":"Order 1042 checkout completed"},{"id":"fulfillment-1042","order_id":"1042","kind":"fulfillment","customer":"c-7","text":"Order 1042 left the warehouse"},{"id":"receipt-1042","order_id":"1042","kind":"receipt","customer":"c-7","text":"Receipt for order 1042 is available"},{"id":"update-1042","order_id":"1042","kind":"customer_update","customer":"c-7","text":"Customer was notified that order 1042 shipped"}]'

curl --fail-with-body --get http://localhost:8080/search \
  --data-urlencode 'q=where can I find proof of payment' \
  --data-urlencode 'kind=receipt' \
  --data-urlencode 'top_k=3'
