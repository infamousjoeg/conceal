# Threat Model

Conceal relies on the host operating system for storage encryption and access
control. Secrets are never written to disk by the application. Attackers with
access to the user's session could retrieve secrets via the OS APIs.

Ensure the underlying OS secret store is protected with user authentication
and follows vendor best practices.
