# balance-api-service

# Build

    git clone https://github.com/udhos/balance-api-service
    cd balance-api-service
    go test ./balance-service
    go install ./balance-service

# Example for A10 device

See sample shell scripts in directory 'samples' for API recipes using 'curl'.

    cd samples                                                   ;# enter samples directory

    export AUTH=admin:a10                                        ;# set A10 device credentials
    export URL=http://localhost:8080/v1/at2/node/1.1.1.1/backend ;# 1.1.1.1 is IP address for A10 device

    ./server_list.sh    ;# list servers
    ./server_create.sh  ;# create server
    ./server_delete.sh  ;# delete server
    ./server_link.sh    ;# link server to parent service group
    ./server_unlink.sh  ;# unlink server from parent service group

# Recipe forward for F5

    curl -sku admin:admin https://1.1.1.1/mgmt/tm/ltm/virtual/ | jq | less
    curl -sku admin:admin https://1.1.1.1/mgmt/tm/ltm/pool/ | jq | less
    curl -sku admin:admin https://1.1.1.1/mgmt/tm/ltm/node/ | jq | less

# Reference

https://devcentral.f5.com/wiki/iControlREST.HomePage.ashx

# HTTP Methods

https://stackoverflow.com/questions/630453/put-vs-post-in-rest

- POST to a URL creates a child resource at a server defined URL.
- PUT to a URL creates/replaces the resource in its entirety at the client defined URL.
- PATCH to a URL updates part of the resource at that client defined URL.

