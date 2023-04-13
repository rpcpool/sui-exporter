#!/bin/bash

curl "${@}" -H "Content-Type: application/json" --data '{"method": "suix_getLatestSuiSystemState", "jsonrpc": "2.0", "id": 123}'
