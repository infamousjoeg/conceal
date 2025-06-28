# Threat Model

Conceal relies on the host operating system for storage encryption and access
control. Secrets are never written to disk by the application. Attackers with
access to the user's session could retrieve secrets via the OS APIs.

Ensure the underlying OS secret store is protected with user authentication
and follows vendor best practices.

## Least Privilege

Only the Conceal binary should have access to the secrets it manages. Users
and automated identities should be granted the minimal permissions necessary
to store or retrieve credentials. Limiting access rights reduces the impact if
an account or process is compromised.

## Endpoint Security

Because Conceal depends on operating system APIs, the host must remain
secure. Employ endpoint protection such as anti‑malware tooling and regular
patch management to prevent privilege escalation or memory scraping attacks
that could bypass OS keyring protections.
