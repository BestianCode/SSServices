# Simple SquidHelper for Update IP-address of user in SQL database

### ENG:
* Update IP-address of user in SQL database

### RUS:
* Обновляет IP адрес пользователя в SQL базе данных

example squid.conf:
-------------------

external_acl_type ip_users ttl=120 children-startup=5 children-max=50 %SRC %LOGIN /server/SSS/SquidHelperAD/SquidHelperDBUpdateIP.gl -config=/server/SSS/SquidHelperAD/SquidHelperAD.json

acl ip_users external ip_users

http_access deny !ip_users
http_access allow auth_users
http_access deny all
