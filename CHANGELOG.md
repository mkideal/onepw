CHANGELOG
=========

# v0.2.0

* Using scrypt generates key for AES-256 instead of md5sum(you **MUST** upgrade password by `onepw up`).

# v0.1.0

* Fix master account(random account, and you **MUST** upgrade password by `onepw up`).
* Replace env varible PASSWORD_MASTER with ONEPW_MASTER, and add new env varible: ONEPW_FILE
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
