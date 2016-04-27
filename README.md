# onepw

onepw is a command line tool for managing passwords

Usage: onepw <COMMAND> [OPTIONS]

Basic workflow:

	#1. init, create file password.data
	$> onepw init

	#2. add a new password
	$> onepw add --label=email -u user@example.com
	type the password:
	repeat the password:

	#3. list all passwords
	$> onepw list

	#optional
	# upload cloud(e.g. dropbox or github or bitbucket

Options:

  -h, --help
      display help

  -v, --version
      display version information

  --master[=$PASSWORD_MASTER]
      master password

Commands:
  help      display help
  version   display version
  init      init password box or modify master password
  add       add a new password or update old password
  remove    remove passwords
  list      list all passwords
