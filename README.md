# onepw [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/mkideal/onepw/master/LICENSE)

## Install

```
go get github.com/mkideal/onepw
```

## What's this
onepw is a command line tool for managing passwords, provide `init`,`add`,`remove`,`list`,`find` commands. You **MUST** remember the `master password`, and don't tell anyone!

## Commands

1. First of all, you should `init` a password box with master password, the master password can be set by ENV(e.g. echo "export PASSWORD_MASTER=MySecret" >> ~/.bashrc)

```shell
onepw init # master password setted by ENV
```

Or
```shell
onepw init --master=MySecret
```

2. And then, `add` a new password

```shell
$> onepw add -c=email -u user@example.com
type the password:		# enter in terminal
repeat the password:	# enter in terminal, too
```

3. `list` all passwords
```shell
$> onepw list
```

Or
```shell
$> onepw ls
```

4. `remove` passwords by id or account
```shell
$> onepw rm <id1 [id2...]> [--all | -a]
```

5. `find` passwords by id,category,account,...
```shell
$> onepw find <WORD>
```

6. You can use dropbox or bitbucket store passwords

## Example

```shell
$> mkdir mypasswords
$> cd mypasswords
$> git init
$> onepw init
$> onepw add -c email -u user@example.com
type the password: 
repeat the password: 
password d9437f07af7c8b035a4fa9513ace449f updated
$> onepw ls
+---------+----------+------------------+----------+---------------------------+
| ID      | CATEGORY | ACCOUNT          | PASSWORD | UPDATED_AT                |
+---------+----------+------------------+----------+---------------------------+
| d9437f0 | email    | user@example.com | 123456   | 2016-04-29T00:54:36+08:00 |
+---------+----------+------------------+----------+---------------------------+
$> onepw add -c github -u hello --pw=123456 --cpw=123456
password 3439d3178f35f56f4c3d6f27e7ccc9a7 updated
$> onepw ls
+---------+----------+------------------+----------+---------------------------+
| ID      | CATEGORY | ACCOUNT          | PASSWORD | UPDATED_AT                |
+---------+----------+------------------+----------+---------------------------+
| 3439d31 | github   | hello            | 123456   | 2016-04-29T00:56:26+08:00 |
+---------+----------+------------------+----------+---------------------------+
| d9437f0 | email    | user@example.com | 123456   | 2016-04-29T00:54:36+08:00 |
+---------+----------+------------------+----------+---------------------------+
$> onepw add -c email -u user2@gmail.com --site=gmail.com --tag=google
type the password:
repeat the password:
password 2ca000f993a665337bebd4700cfd7c6c updated
$> onepw ls
+---------+----------+------------------+----------+---------------------------+
| ID      | CATEGORY | ACCOUNT          | PASSWORD | UPDATED_AT                |
+---------+----------+------------------+----------+---------------------------+
| 2ca000f | email    | user2@gmail.com  | 123456   | 2016-04-29T00:58:49+08:00 |
+---------+----------+------------------+----------+---------------------------+
| 3439d31 | github   | hello            | 123456   | 2016-04-29T00:56:26+08:00 |
+---------+----------+------------------+----------+---------------------------+
| d9437f0 | email    | user@example.com | 123456   | 2016-04-29T00:54:36+08:00 |
+---------+----------+------------------+----------+---------------------------+
$> onepw find mail
+---------+-------+------------------+--------+---------------------------+
| 2ca000f | email | user2@gmail.com  | 123456 | 2016-04-29T00:58:49+08:00 |
+---------+-------+------------------+--------+---------------------------+
| d9437f0 | email | user@example.com | 123456 | 2016-04-29T00:54:36+08:00 |
+---------+-------+------------------+--------+---------------------------+
$> onepw find hello
+---------+--------+-------+--------+---------------------------+
| 3439d31 | github | hello | 123456 | 2016-04-29T00:56:26+08:00 |
+---------+--------+-------+--------+---------------------------+
$> onepw rm 343
deleted passwords:
3439d3178f35f56f4c3d6f27e7ccc9a7
$> onepw ls
+---------+----------+------------------+----------+---------------------------+
| ID      | CATEGORY | ACCOUNT          | PASSWORD | UPDATED_AT                |
+---------+----------+------------------+----------+---------------------------+
| 2ca000f | email    | user2@gmail.com  | 123456   | 2016-04-29T00:58:49+08:00 |
+---------+----------+------------------+----------+---------------------------+
| d9437f0 | email    | user@example.com | 123456   | 2016-04-29T00:54:36+08:00 |
+---------+----------+------------------+----------+---------------------------+
```
