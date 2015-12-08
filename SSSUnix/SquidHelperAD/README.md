# Simple SquidHelper for ActiveDirectory

### ENG:
* Allows authorization in multiple Active Directory domains.
* Authorization in Active Directory through LDAP. Requires access to tcp port 389 on domain controllers.
* Temporary not supported NTLM, and Kerberos. Only basic authorization with the username and password.
* May check login and password in SQL database.

### RUS:
* Позволяет авторизоваться в нескольких доменах Active  Directory.
* Авторизация в Active Directory происходит через LDAP. Необходим доступ к 389 порту контроллеров домена.
* Пока не поддерживает NTLM и Kerbberos. Только облычная авторизация с вводом логина и пароля.
* Также может проверять логин и пароль в SQL базе.

example squid.conf:
-------------------

auth_param basic program /server/SSS/SquidHelperAD/SquidHelperAD.gl -config=/server/SSS/SquidHelperAD/SquidHelperAD.json

auth_param basic children 50 startup=5 idle=1

