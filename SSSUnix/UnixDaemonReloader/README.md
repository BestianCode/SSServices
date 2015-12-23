# UnixDaemonReloader

### ENG:
* Automatic restart unix-demons when the configuration files are modified.
* Execution any commands when changing specified files.

### RUS:
* Автоматическая перезагрузка unix-демонов при изменении конфигурационного файла.
* Выполнение любых команд при изменении указанных файлов.

#### *Default install folder: /server/SSS/UnixDaemonReloader*

# Configuration file syntax:
* _Шелл, и ключ к нему, для запуска внешних команд в таком виде: /bin/sh -c "ps ax"_
* _Path to Unix shell and the parameter allows you to execute an external command, like so: /bin/sh -c "ps ax"_
#####	"UDR_Shell":		"/bin/sh",
#####	"UDR_ShellExecParam":	"-c",

* _в виде списка задаем строки с отслеживаемыми файлами и каталогами_
* "/каталог", "файл", "действие",
* "/каталог", "маска*файла*", "действие",
* "/каталог", "!все*файлы*кроме*этого,!кроме*этого,!и*кроме*этого", "действие"

* _specify a list of strings to track files and directories_
* "/directory", "file", "action",
* "/directory", "mask*of*the*files*", "action",
* "/directory", "!all*files*except*this,!except*this,!and*except*this", "action"
#####	"UDR_WatchList":		[
#####				["/etc/postfix", "main.cf", "/etc/init.d/postfix reload"],
#####				["/etc/postfix", "master.cf", "/etc/init.d/postfix reload"],
#####				["/etc/amavis/conf.d", "*", "/etc/init.d/amavis restart"],
#####				["/etc/spamassassin", "local.cf", "/etc/init.d/spamassassin reload"],
#####				["/etc","iptables.conf","/server/scripts/iptables/rest.sh"]
#####					],

* _пауза в секундах перед запуском скрипта. Этот параметр сделан для того, что бы если вы вдруг случайно во время редактирования конфига сохранили файл "недоделанным", то у вас было время на исправление ошибки до перезапуска демона._
* _pause before running the script (seconds). This setting for save your daemons from "your hands". If you during editing configuration file, accidentally press "save a file" with error or unfinished, then you have time to correct the error before the daemon will be restarted._
#####	"UDR_PauseBefore":	600,

* _сколько спать между циклами проверки конфигов_
* _How much time to sleep between checks files_
#####	"Sleep_Time":		60,

* _путь к базе SQLite, в которой хранятся контрольные суммы файлов_
* _SQLite_DB - the way to the base SQLite, which stores the checksums of files_
#####	"SQLite_DB":		"/server/SSS/UnixDaemonReloader/UnixDaemonReloader.sqlite",

* PID, LOG and LOG Level :)
#####	"PID_File":		"/var/run/SSS/UnixDaemonReloader.pid",
#####	"LOG_File":		"/var/log/SSS/UnixDaemonReloader.log",
#####	"LOG_Level":		0
