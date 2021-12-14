package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Mistat/ldap-pg"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	version  string
	revision string

	fs         = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	dbHostName = fs.String(
		"h",
		"localhost",
		"DB Hostname",
	)
	dbPort = fs.Int(
		"p",
		5432,
		"DB Port",
	)
	dbName = fs.String(
		"d",
		"",
		"DB Name",
	)
	dbSchema = fs.String(
		"s",
		"",
		"DB Schema",
	)
	dbUser = fs.String(
		"u",
		"",
		"DB User",
	)
	dbPassword = fs.String(
		"w",
		"",
		"DB Password",
	)
	dbMaxOpenConns = fs.Int(
		"db-max-open-conns",
		5,
		"DB max open connections",
	)
	dbMaxIdleConns = fs.Int(
		"db-max-idle-conns",
		2,
		"DB max idle connections",
	)
	suffix = fs.String(
		"suffix",
		"",
		"Suffix for the LDAP",
	)
	rootdn = fs.String(
		"root-dn",
		"",
		"Root dn for the LDAP",
	)
	rootpw = fs.String(
		"root-pw",
		"",
		"Root password for the LDAP",
	)
	bindAddress = fs.String(
		"b",
		"127.0.0.1:8389",
		"Bind address",
	)
	logLevel = fs.String(
		"log-level",
		"info",
		"Log level, on of: debug, info, warn, error, alert",
	)
	pprofServer = fs.String(
		"pprof",
		"",
		"Bind address of pprof server (Don't start the server with default)",
	)
	gomaxprocs = fs.Int(
		"gomaxprocs",
		0,
		"GOMAXPROCS (Use CPU num with default)",
	)
	passThroughLDAPDomain = fs.String(
		"pass-through-ldap-domain",
		"",
		"Pass-through/LDAP: Domain for pass-through/LDAP",
	)
	passThroughLDAPServer = fs.String(
		"pass-through-ldap-server",
		"",
		"Pass-through/LDAP: Server address and port (e.g. myldap:389)",
	)
	passThroughLDAPSearchBase = fs.String(
		"pass-through-ldap-search-base",
		"",
		"Pass-through/LDAP: Search base",
	)
	passThroughLDAPFilter = fs.String(
		"pass-through-ldap-filter",
		"",
		"Pass-through/LDAP: Filter for finding an user (e.g. (cn=%u))",
	)
	passThroughLDAPBindDN = fs.String(
		"pass-through-ldap-bind-dn",
		"",
		"Pass-through/LDAP: Bind DN",
	)
	passThroughLDAPPassword = fs.String(
		"pass-through-ldap-password",
		"",
		"Pass-through/LDAP: Bind password",
	)
	passThroughLDAPScope = fs.String(
		"pass-through-ldap-scope",
		"sub",
		"Pass-through/LDAP: Search scope, on of: base, one, sub",
	)
	passThroughLDAPTimeout = fs.Int(
		"pass-through-ldap-timeout",
		10,
		"Pass-through/LDAP: Timeout seconds",
	)
	migrationEnabled = fs.Bool(
		"migration",
		false,
		"Enable migration mode which means LDAP server accepts add/modify operational attributes (default false)",
	)
	defaultPPolicyDN = fs.String(
		"default-ppolicy-dn",
		"",
		"DN of the default password policy entry (e.g. cn=standard-policy,ou=Policies,dc=example,dc=com)",
	)
)

func main() {
	fs.Var(&ldap_pg.CustomSchema, "schema", "Additional/overwriting custom schema")

	var aclFlags ldap_pg.ArrayFlags
	fs.Var(&aclFlags, "acl", `Simple ACL: the format is <DN(User, Group or empty(everyone))>:<Scope(R, W or RW)>:<Invisible Attributes> (e.g. cn=reader,dc=example,dc=com:R:userPassword,telephoneNumber)`)

	fmt.Fprintf(os.Stdout, "ldap-pg %s (rev: %s)\n", version, revision)
	fs.Usage = func() {
		_, exe := filepath.Split(os.Args[0])
		fmt.Fprintf(os.Stderr, "\nUsage:\n\n  %s [options]\n\nOptions:\n\n", exe)
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error: Cannot parse the args: %v, err: %s", os.Args[1:], err)
	}

	if len(os.Args) == 1 {
		fs.Usage()
		return
	}

	rootPW := *rootpw
	*rootpw = ""

	passThroughConfig := &ldap_pg.PassThroughConfig{}
	if *passThroughLDAPDomain != "" {
		passThroughConfig.Add(*passThroughLDAPDomain, &ldap_pg.LDAPPassThroughClient{
			Server:     *passThroughLDAPServer,
			SearchBase: *passThroughLDAPSearchBase,
			Timeout:    *passThroughLDAPTimeout,
			Filter:     *passThroughLDAPFilter,
			BindDN:     *passThroughLDAPBindDN,
			Password:   *passThroughLDAPPassword,
			Scope:      *passThroughLDAPScope,
		})
	}

	var acl []string
	if aclFlags != nil {
		acl = strings.Split(aclFlags.String(), "\n")
	}

	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := ldap_pg.NewServer(&ldap_pg.ServerConfig{
		DBHostName:        *dbHostName,
		DBPort:            *dbPort,
		DBName:            *dbName,
		DBSchema:          *dbSchema,
		DBUser:            *dbUser,
		DBPassword:        *dbPassword,
		DBMaxOpenConns:    *dbMaxOpenConns,
		DBMaxIdleConns:    *dbMaxIdleConns,
		Suffix:            *suffix,
		RootDN:            *rootdn,
		RootPW:            rootPW,
		BindAddress:       *bindAddress,
		PassThroughConfig: passThroughConfig,
		LogLevel:          *logLevel,
		PProfServer:       *pprofServer,
		GoMaxProcs:        *gomaxprocs,
		MigrationEnabled:  *migrationEnabled,
		QueryTranslator:   "default",
		SimpleACL:         acl,
		DefaultPPolicyDN:  *defaultPPolicyDN,
	})

	go server.Start(*bindAddress)

	<-ctx.Done()
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("info: Shutdown ldap-pg...")
	server.Stop()
}
