# balance-api-service

# Build

    git clone https://github.com/udhos/balance-api-service
    cd balance-api-service
    go test ./balance-service
    go install ./balance-service

# Test Recipe

Backend API

    # backend API
    #
    # delete backend:
    curl -u admin:a10 -X DELETE -d '{"BackendName": "eraseme1"}' http://192.168.56.20:8080/v1/at2/node/10.255.255.6/backend
    # unlink all backend ports from service group:
    curl -u admin:a10 -X DELETE -d '{"BackendName": "eraseme1", "ServiceGroups": [{"Name": "group1"}]}' http://192.168.56.20:8080/v1/at2/node/10.255.255.6/backend

Caution: rule API below is broken

    # rule API
    curl -u admin:a10 -X PUT --data-binary '@sample.txt' http://192.168.56.20:8080/v1/at2/node/10.255.255.6/rule
    curl -u admin:a10 -X PUT -d '[]' http://192.168.56.20:8080/v1/at2/node/10.255.255.6/rule
    curl -k -u admin:a10 http://192.168.56.20:8080/v1/at2/node/10.255.255.6/rule
    curl -k -u admin:a10 http://192.168.56.20:8080/v1/at3/node/10.255.255.6/rule

# Recipe forward for F5

    curl -sku admin:admin https://10.255.255.120/mgmt/tm/ltm/virtual/ | jq | less
    curl -sku admin:admin https://10.255.255.120/mgmt/tm/ltm/pool/ | jq | less
    curl -sku admin:admin https://10.255.255.120/mgmt/tm/ltm/node/ | jq | less

# Reference

https://devcentral.f5.com/wiki/iControlREST.HomePage.ashx

# HTTP Methods

https://stackoverflow.com/questions/630453/put-vs-post-in-rest

- POST to a URL creates a child resource at a server defined URL.
- PUT to a URL creates/replaces the resource in its entirety at the client defined URL.
- PATCH to a URL updates part of the resource at that client defined URL.

