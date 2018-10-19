# f5-api-service

# Recipe

    curl -sku admin:admin https://10.255.255.120/mgmt/tm/ltm/virtual/ | jq | less
    curl -sku admin:admin https://10.255.255.120/mgmt/tm/ltm/pool/ | jq | less
    curl -sku admin:admin https://10.255.255.120/mgmt/tm/ltm/node/ | jq | less

# Reference

https://devcentral.f5.com/wiki/iControlREST.HomePage.ashx

# Methods

https://stackoverflow.com/questions/630453/put-vs-post-in-rest

- POST to a URL creates a child resource at a server defined URL.
- PUT to a URL creates/replaces the resource in its entirety at the client defined URL.
- PATCH to a URL updates part of the resource at that client defined URL.

