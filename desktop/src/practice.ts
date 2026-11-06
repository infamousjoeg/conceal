import type { Client, Inbox, Reply, Review, RunResult } from "./model";

const ok = <T,>(data: T): Reply<T> => ({ data: structuredClone(data), error: null });
const failure = <T,>(error: string): Reply<T> => ({ data: null, error });

export class PracticeClient implements Client {
  private requests = new Map<string, Review>();
  constructor() { this.reset(); }
  reset() {
    this.requests.clear();
    const now = new Date().toISOString();
    for (const s3 of [false, true]) {
      const id = (s3 ? "b" : "a").repeat(32);
      const digest = (s3 ? "2" : "1").repeat(64);
      const fields = s3 ? [
        { label: "Bucket", value: "example-backups" }, { label: "AWS region", value: "us-east-1" },
        { label: "Prefix", value: "reports/2026/" }, { label: "Maximum results", value: "10" },
        { label: "Pagination", value: "One page; no automatic follow-up requests" },
      ] : [
        { label: "Repository", value: "acme/website" }, { label: "Title", value: "Fix the broken contact link" },
        { label: "From branch", value: "fix/contact-link" }, { label: "Into branch", value: "main" },
        { label: "Description", value: "Updates the contact link in the footer.\n\nPlease review the change before merging." },
        { label: "Draft", value: "Yes" }, { label: "Maintainer edits", value: "Disabled" },
      ];
      const body = { title: "Fix the broken contact link", head: "fix/contact-link", base: "main",
        body: fields[4].value, draft: true, maintainer_can_modify: false };
      const contract = {
        schema_version: 2, name: s3 ? "backup-inventory" : "website-review", adapter: s3 ? "aws" : "github",
        credentials: s3 ? { access_key_id: "aws/access-key-id", secret_access_key: "aws/secret-access-key", session_token: "aws/session-token" } : { token: "github/token" },
        actions: [s3 ? "aws.s3.objects.list" : "github.pulls.create-draft"], approval_ttl_seconds: 300,
        scope: s3 ? { region: "us-east-1", buckets: [{ name: "example-backups", prefix: "reports/" }] } : { repositories: ["acme/website"] },
      };
      this.requests.set(id, {
        schema_version: 1, proposal: { digest, request: {
          id, created_at: now, contract_path: s3 ? "(practice)/s3-contract.json" : "(practice)/github-contract.json",
          contract, action: contract.actions[0], resource: s3 ? "example-backups" : "acme/website",
          input: s3 ? { prefix: "reports/2026/", limit: 10 } : body,
          http: s3 ? { method: "GET", url: "https://s3.us-east-1.amazonaws.com/example-backups?encoding-type=url&list-type=2&max-keys=10&prefix=reports%2F2026%2F" } :
            { method: "POST", url: "https://api.github.com/repos/acme/website/pulls", body },
          plan: { adapter_version: "1", read_only: s3 },
        }},
        review: {
          title: s3 ? "List S3 object names" : "Create a draft pull request",
          consequence: s3 ? "Returns one page of object names and metadata, not file contents. Names may be private. AWS request charges may apply." :
            "Creates a draft pull request visible to repository collaborators. This does not merge code. GitHub may notify subscribers.",
          fields,
        },
        adapter: s3 ? "aws" : "github", credentials: Object.entries(contract.credentials).map(([label, value]) => ({ label, value: value! })),
        status: "pending", can_approve: true, can_execute: false, can_reject: true, server_time: now,
      });
    }
  }
  async inbox(): Promise<Reply<Inbox>> {
    return ok({ schema_version: 1, total: this.requests.size, truncated: false, items: [...this.requests.values()].map(r => ({
      id: r.proposal.request.id, title: r.review.title, adapter: r.adapter, action: r.proposal.request.action,
      resource: r.proposal.request.resource!, created_at: r.proposal.request.created_at, status: this.current(r).status,
    })) });
  }
  private current(r: Review): Review {
    r.server_time = new Date().toISOString();
    if (r.status === "approved" && r.approval && Date.parse(r.approval.expires_at) <= Date.now()) {
      r.status = "expired"; r.can_execute = false;
    }
    return r;
  }
  async review(id: string): Promise<Reply<Review>> {
    const r = this.requests.get(id);
    return r ? ok(this.current(r)) : failure("Request not found.");
  }
  async approve(id: string, digest: string): Promise<Reply<unknown>> {
    const r = this.requests.get(id);
    if (!r?.can_approve || r.proposal.digest !== digest) return failure("The reviewed request changed. Refresh and review again.");
    r.approval = { digest, approved_at: new Date().toISOString(), expires_at: new Date(Date.now() + 300_000).toISOString() };
    r.status = "approved"; r.can_approve = false; r.can_execute = true;
    return ok(r.approval);
  }
  async reject(id: string): Promise<Reply<unknown>> {
    const r = this.requests.get(id);
    if (!r?.can_reject) return failure("This request can no longer be rejected.");
    r.status = "rejected"; r.can_approve = r.can_execute = r.can_reject = false;
    r.receipt = { id, digest: r.proposal.digest, outcome: "revoked", time: new Date().toISOString() };
    return ok(r.receipt);
  }
  async execute(id: string): Promise<Reply<RunResult>> {
    const r = this.requests.get(id);
    if (!r || !this.current(r).can_execute) return failure("An unexpired approval is required.");
    r.status = "consumed"; r.can_approve = r.can_execute = r.can_reject = false;
    r.receipt = { id, digest: r.proposal.digest, outcome: "succeeded", http_status: 200, time: new Date().toISOString() };
    return ok({ receipt: r.receipt, response: r.adapter === "aws" ?
      { example_objects: ["reports/2026/january.csv", "reports/2026/february.csv"], simulated: true } :
      { number: 42, draft: true, simulated: true } });
  }
}
