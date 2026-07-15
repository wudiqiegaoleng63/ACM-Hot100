# Dependency vulnerability exception policy

Security scans are merge gates. An exception is allowed only when no safe upgrade or mitigation is currently available and the affected path is proven unreachable or otherwise contained.

Every exception must be a dated entry in this file containing:

- GHSA, CVE, or Go vulnerability ID;
- affected package and locked version;
- reachability and impact analysis;
- temporary mitigation;
- owner;
- expiration date no more than 30 days away;
- removal criteria.

Expired exceptions fail review and must not be silently renewed. Broad package, severity, or scanner ignores are forbidden.

## Active exceptions

None.
