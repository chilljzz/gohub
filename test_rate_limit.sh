#!/bin/bash


for i in {1..20}
do

curl -s \
-o /dev/null \
-w "%{http_code}\n" \
-X POST \
http://localhost:8081/api/users/login \
-H "Content-Type: application/json" \
-d '
{
    "username":"test",
    "password":"wrong-password"
}
'

done
