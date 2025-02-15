package ldap_pg

import (
	"context"
	"log"

	ldap "github.com/openstandia/ldapserver"
	"golang.org/x/xerrors"
)

func handleModifyDN(s *Server, w ldap.ResponseWriter, m *ldap.Message) {
	ctx := SetSessionContext(context.Background(), m)

	r := m.GetModifyDNRequest()
	dn, err := s.NormalizeDN(string(r.Entry()))

	if err != nil {
		log.Printf("warn: Invalid dn: %s err: %s", r.Entry(), err)

		// TODO return correct error
		responseModifyDNError(w, err)
		return
	}

	if !s.RequiredAuthz(m, ModRDNOps, dn) {
		responseModifyDNError(w, NewInsufficientAccess())
		return
	}

	newDN, oldRDN, err := dn.ModifyRDN(s.schemaMap, string(r.NewRDN()), bool(r.DeleteOldRDN()))

	if err != nil {
		// TODO return correct error
		log.Printf("info: Invalid newrdn. dn: %s newrdn: %s err: %#v", dn.DNNormStr(), r.NewRDN(), err)
		responseModifyDNError(w, err)
		return
	}

	log.Printf("info: Modify DN entry: %s", dn.DNNormStr())

	if r.NewSuperior() != nil {
		sup := string(*r.NewSuperior())
		newParentDN, err := s.NormalizeDN(sup)
		if err != nil {
			// TODO return correct error
			responseModifyDNError(w, NewInvalidDNSyntax())
			return
		}

		newDN, err = newDN.Move(newParentDN)
		if err != nil {
			// TODO return correct error
			responseModifyDNError(w, NewInvalidDNSyntax())
			return
		}
	}

	i := 0
Retry:

	err = s.Repo().UpdateDN(ctx, dn, newDN, oldRDN)
	if err != nil {
		var retryError *RetryError
		if ok := xerrors.As(err, &retryError); ok {
			if i < maxRetry {
				i++
				log.Printf("warn: Detect consistency error. Do retry. try_count: %d", i)
				goto Retry
			}
			log.Printf("error: Give up to retry. try_count: %d", i)
		}

		log.Printf("warn: Failed to modify dn: %s err: %+v", dn.DNNormStr(), err)
		// TODO error code
		responseModifyDNError(w, err)
		return
	}

	res := ldap.NewModifyDNResponse(ldap.LDAPResultSuccess)
	w.Write(res)
}

func responseModifyDNError(w ldap.ResponseWriter, err error) {
	var ldapErr *LDAPError
	if ok := xerrors.As(err, &ldapErr); ok {
		log.Printf("warn: ModifyDN LDAP error. err: %+v", err)

		res := ldap.NewModifyDNResponse(ldapErr.Code)
		if ldapErr.Msg != "" {
			res.SetDiagnosticMessage(ldapErr.Msg)
		}
		w.Write(res)
	} else {
		log.Printf("error: ModifyDN error. err: %+v", err)

		// TODO
		res := ldap.NewModifyDNResponse(ldap.LDAPResultProtocolError)
		w.Write(res)
	}
}
