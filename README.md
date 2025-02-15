# ldap-pg

[![GoDoc](https://godoc.org/github.com/openstandia/ldap-pg?status.svg)](https://godoc.org/github.com/openstandia/ldap-pg)

**This repository is heavily under development.**

**ldap-pg** is a LDAP server implementation which uses PostgreSQL as the backend database.

## Features

- Basic LDAP operations
  - Bind
    - [x] PLAIN
    - [x] SSHA
    - [x] SSHA256
    - [x] SSHA512
    - [x] ARGON2
    - [x] Pass-through authentication (Support `{SASL}foo@domain` format)
  - Search
    - [x] base
    - [x] one
    - [x] sub
    - [x] children
  - [x] Add
  - [x] Modify
  - [x] Delete
  - ModifyDN
    - [x] Rename RDN
    - [x] Support deleteoldrdn
    - [x] Support newsuperior
  - [ ] Compare
  - [ ] Extended
- LDAP Controls
  - [x] Simple Paged Results Control
  - [ ] Sort Control
- Support association (like OpenLDAP memberOf overlay)
  - [x] Return memberOf attribute as operational attribute
  - [x] Maintain member uniqueMember / memberOf
  - [x] Search filter using memberOf
- Schema
  - [x] Basic schema processing
  - [ ] More schema processing
  - [x] User defined schema
  - [ ] Multiple RDNs
- Password Policy
  - [x] Account lock
  - [ ] More policy controls
- Authorization
  - [x] Simple ACL
- Last bind
  - [x] Record the timestamp of the last successful bind
- Network
  - [ ] SSL/StartTLS
- [ ] Prometheus metrics
- [x] Auto create table for PostgreSQL
- [ ] Auto migrate table for PostgreSQL

## Requirement

PostgreSQL 12 or later.

## Install

### From binary

Please download it from [release page](../../releases).

### From source

`ldap-pg` is written by Go. Install Go then build `ldap-pg`:

```
make
```

You can find the binary in `./bin/` directory.

## Usage

### Start `ldap-pg`

```
ldap-pg.

Usage:

  ldap-pg [options]

Options:

  -acl value
        Simple ACL: the format is <DN(User, Group or empty(everyone))>:<Scope(R, W or RW)>:<Invisible Attributes> (e.g. cn=reader,dc=example,dc=com:R:userPassword,telephoneNumber)
  -b string
        Bind address (default "127.0.0.1:8389")
  -d string
        DB Name
  -db-max-idle-conns int
        DB max idle connections (default 2)
  -db-max-open-conns int
        DB max open connections (default 5)
  -default-ppolicy-dn string
        DN of the default password policy entry (e.g. cn=standard-policy,ou=Policies,dc=example,dc=com)
  -gomaxprocs int
        GOMAXPROCS (Use CPU num with default)
  -h string
        DB Hostname (default "localhost")
  -log-level string
        Log level, on of: debug, info, warn, error, alert (default "info")
  -migration
        Enable migration mode which means LDAP server accepts add/modify operational attributes (default false)
  -p int
        DB Port (default 5432)
        If use unix socket set "0"
  -pass-through-ldap-bind-dn string
        Pass-through/LDAP: Bind DN
  -pass-through-ldap-domain string
        Pass-through/LDAP: Domain for pass-through/LDAP
  -pass-through-ldap-filter string
        Pass-through/LDAP: Filter for finding an user (e.g. (cn=%u))
  -pass-through-ldap-password string
        Pass-through/LDAP: Bind password
  -pass-through-ldap-scope string
        Pass-through/LDAP: Search scope, on of: base, one, sub (default "sub")
  -pass-through-ldap-search-base string
        Pass-through/LDAP: Search base
  -pass-through-ldap-server string
        Pass-through/LDAP: Server address and port (e.g. myldap:389)
  -pass-through-ldap-timeout int
        Pass-through/LDAP: Timeout seconds (default 10)
  -pprof string
        Bind address of pprof server (Don't start the server with default)
  -root-dn string
        Root dn for the LDAP
  -root-pw string
        Root password for the LDAP
  -s string
        DB Schema
  -schema value
        Additional/overwriting custom schema
  -suffix string
        Suffix for the LDAP
  -u string
        DB User
  -w string
        DB Password

```

#### Example

Start PostgreSQL server.

```
docker run --rm -d \
  -e POSTGRES_DB=testdb \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_PASSWORD=testpass \
  -p 35432:5432 \
  postgres:13-alpine \
  -c log_destination=stderr \
  -c log_statement=all \
  -c log_connections=on \
  -c log_disconnections=on \
  -c jit=off
```

Start `ldap-pg` server.

```
ldap-pg -h localhost -u testuser -w testpass -d testdb -s public \
 -suffix dc=example,dc=com -root-dn cn=Manager,dc=example,dc=com -root-pw secret \
 -log-level info

[  info ] 2019/10/03 15:13:37 main.go:169: Setup GOMAXPROCS with NumCPU: 8
[  info ] 2019/10/03 15:13:37 main.go:234: Starting ldap-pg on 127.0.0.1:8389
```

`ldap-pg` creates required tables and indexes into the PostgreSQL if not exists.
You can import your LDIF file by using standard LDAP tools like `ldapadd` command.

```
$ cat << EOS > base.ldif

dn: dc=example,dc=com
objectClass: top
objectClass: dcObject
objectClass: organization
o: Example Inc.
dc: example

dn: ou=Users,dc=example,dc=com
objectClass: organizationalUnit
ou: Users

dn: ou=Groups,dc=example,dc=com
objectClass: organizationalUnit
ou: Group

EOS

$ ldapadd -H ldap://localhost:8389 -x -D cn=manager,dc=example,dc=com -w secret -f base.ldif
adding new entry "dc=example,dc=com"

adding new entry "ou=Users,dc=example,dc=com"

adding new entry "ou=Groups,dc=example,dc=com"
```

## Integration Test

Start PostgreSQL server.

```
docker run --rm -d \
  -e POSTGRES_DB=testdb \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_PASSWORD=testpass \
  -p 35432:5432 \
  postgres:13-alpine \
  -c log_destination=stderr \
  -c log_statement=all \
  -c log_connections=on \
  -c log_disconnections=on \
  -c jit=off
```

Run test cases.

```
make it
```

## License

Licensed under the [GPL](/LICENSE) license.
