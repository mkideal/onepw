CHANGELOG
=========

# HEAD

* Add `-p(--just-password)` and `-f(--just-first)` flags for find command

# v0.0.2

* Add validating master password(you **SHOULD** upgrade password.data by `onepw up`)
* Add command `upgrade` (aliases `up`)
* Add secret prompt for typing password

# v0.0.1

* First version of onepw, contains following features
* Supported commands: `help`,`version`,`init`,`add`,`remove`,`list`,`find`
* Encrypted by CFB mode with AES-256
* Each password contains Category,Account,Password,Site,Tags,Ext,CreatedAt,LastUpdatedAt.
