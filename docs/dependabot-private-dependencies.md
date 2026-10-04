# Dependabot and the private gomedz dependency

## Failure identified

Dependabot can clone medzoner-go and resolve public module versions, but cannot
read `github.com/medzoner-org/gomedz`. Its proxy reports a failed JIT access
request (403), then `git_dependencies_not_reachable`.

The public Go proxy's 404 for a private module is expected. Changing Go versions,
GOPRIVATE in an Actions job, runner timeouts or Fiber itself cannot grant
Dependabot access to another private repository.

## Preferred fix: organization-native access

Because the consumer and dependency are both in medzoner-org, no PAT is required
if an organization owner grants Dependabot access to the dependency repository:

1. Open [medzoner-org settings](https://github.com/organizations/medzoner-org/settings).
2. In the security section, open **Advanced Security → Global settings**
   (the labels may vary with the GitHub UI).
3. Find **Grant Dependabot access to private repositories**.
4. Find and select **gomedz**, the repository containing the dependency.
   Merely selecting medzoner-go does not grant access to its dependency.
5. Retry the failed update from medzoner-go's Dependabot updates page, or trigger
   a fresh check after the permission is saved.

GitHub warns that granting this access allows organization users to access
repository contents through Dependabot updates. Select only the shared library,
not every private repository, and review that scope before enabling it.

This administrative setting cannot be changed by the repository's CI YAML.
Dependabot security updates are already enabled; this fix does not require a
new version-update schedule or a committed dependabot.yml.

Reference: [GitHub documentation on private repository access](https://docs.github.com/en/code-security/how-tos/secure-at-scale/configure-organization-security/establish-complete-coverage/configure-global-settings#granting-dependabot-access-to-private-repositories).

## Alternative: an explicit Git registry

If organization-native access is not available, use a fine-grained PAT with
read-only Contents access to gomedz, owned by an account authorized to read it.
Save it as a **Dependabot** secret named `GOMEDZ_READ_TOKEN`, not only an Actions
secret. An organization secret must be shared with the consumer medzoner-go.
Never paste the token into a configuration file, issue, PR or log.

Then create `.github/dependabot.yml`, for example:

```yaml
version: 2
registries:
  gomedz-git:
    type: git
    url: https://github.com
    username: x-access-token
    password: ${{ secrets.GOMEDZ_READ_TOKEN }}
updates:
  - package-ecosystem: gomod
    directory: /
    registries:
      - gomedz-git
    schedule:
      interval: weekly
```

This alternative also enables weekly version updates. It is an example, not an
active configuration: do not deploy it without first creating the secret. Use
one access strategy deliberately rather than adding an unused credential.

Reference: [GitHub documentation on registry credentials and Dependabot secrets](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/configuring-access-to-private-registries-for-dependabot).

## Verification

- A new job must resolve gomedz without JIT 403 / `git_dependencies_not_reachable`.
- Seeing public registry downloads succeed is not proof private access works.
- A merged Fiber update can resolve that alert, but does not prove future
  Dependabot updates are authorized.

The manual Fiber update accompanying this document pins v2.52.15. No security
alerts are disabled, ignored or dismissed as part of the access fix.
