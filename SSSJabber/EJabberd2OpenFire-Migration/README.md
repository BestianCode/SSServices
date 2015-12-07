# Tool for migration from EJabberd to OpenFire

## Step 1
### run: "ejabberdctl --node ejabberd@jabber.mydomain.org export2odbc jabber.mydomain.org ."

* Trying to export Mnesia table 'last' on Host 'jabber.mydomain.org' to file './last.txt'
*   Result: {atomic,ok}
* Trying to export Mnesia table 'offline' on Host 'jabber.mydomain.org' to file './offline.txt'
*   Result: {atomic,ok}
* Trying to export Mnesia table 'private_storage' on Host 'jabber.mydomain.org' to file './private_storage.txt'
*   Result: {atomic,ok}
* Trying to export Mnesia table 'roster' on Host 'jabber.mydomain.org' to file './roster.txt'
*   Result: {atomic,ok}
* Trying to export Mnesia table 'vcard' on Host 'jabber.mydomain.org' to file './vcard.txt'
*   Result: {atomic,ok}
* Trying to export Mnesia table 'vcard_search' on Host 'jabber.mydomain.org' to file './vcard_search.txt'
*   Result: {atomic,ok}
* Trying to export Mnesia table 'passwd' on Host 'jabber.mydomain.org' to file './passwd.txt'
*   Result: {atomic,ok}

## Step 2
### move all *.txt files to directory which contains ./EJabberd2OpenFire.gl to subdir ./files/ (_ex: .../go/src/github.com/BestianRU/SSServices/SSSJabber/EJabberd2OpenFire-Migration/files/_)
### cd .../go/src/github.com/BestianRU/SSServices/SSSJabber/EJabberd2OpenFire-Migration/

## Step 3
* _parse *.txt files and insert all data to SQL DataBase (PostgreSQL, MySQL or SQLite)__
* MySQL - not tested.
* SQLite - too slow.
* PostgreSQL - is best choice.
* **run: ./EJabberd2OpenFire.gl -phase 1**

## Step 4
* I use this phase in my corporation for get info about fullname persons from ERP system for jabber VCard.
* This phase do not working for others without additional programming.
* **Skip this step !!!**
* **run: ./EJabberd2OpenFire.gl -phase 2**

## Step 5
* You can execute manual scripts for update SQL nick/fullname table
* **(_ex: .../go/src/github.com/BestianRU/SSServices/SSSJabber/EJabberd2OpenFire-Migration/files/_manual_UserInfo_update.sql.sample_)**

## Step 6
* Making XML file for import to OpenFire. (Need OpenFire plugin "import/Export")
* **run: ./EJabberd2OpenFire.gl -phase 3**

## Step 7
* **You need to import XML file into OpenFire**

## Step 8
* Inserting VCard data into OpenFire DataBase. (MySQL or PostgreSQL)
* **run: ./EJabberd2OpenFire.gl -phase 4**

#Congratulation! :)

Bonus:
* If you use cluster, copy "Certificates vault" _/etc/openfire/security_ from first server to second and etc.
* If you user cluster with linux/BSD UCarp/Carp, to edit /var/lib/openfire/plugins/hazelcast/classes/hazelcast-cache-config.xml for listening only on non-carp ip. If you do not make it, your cluster will die whenever one of the nodes will go to down.

