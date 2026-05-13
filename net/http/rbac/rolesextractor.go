package rbac

import (
	"net/http"
)

// RolesExtractor pulls the set of roles attached to a request,
// along with an authenticated flag that drives the 401-vs-403
// distinction on denial.
//
// "Roles" is a deliberately broad label: any string the caller wants to
// match against AllowRoles / DenyRoles fits — IAM roles, group
// memberships, OAuth scopes, custom claim values, and so on. The RBAC
// package name commits to role-shaped vocabulary; the extractor stays
// agnostic about where the values come from.
//
// Contract:
//   - roles may be nil or empty for an authenticated caller with no
//     matching roles; the middleware treats len(roles) == 0 as
//     "no positive matches" rather than as unauthenticated.
//   - authenticated reflects whether the request carries a trusted
//     identity at all. An extractor that cannot tell the difference
//     should return true whenever it returns any roles, and false
//     only on a confirmed absence of identity.
//   - the extractor must not mutate the request.
type RolesExtractor func(r *http.Request) (roles []string, authenticated bool)
