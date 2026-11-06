export type Field = { label: string; value: string };
export type Receipt = { id: string; digest: string; outcome: string; http_status?: number; time: string };
export type Review = {
  schema_version: 1;
  proposal: {
    digest: string;
    request: {
      id: string; created_at: string; contract_path: string;
      contract: { name: string; approval_ttl_seconds: number; [key: string]: unknown };
      action: string; repository?: string; resource?: string; input: unknown;
      http: { method: string; url: string; body?: unknown; headers?: Record<string, string> };
      plan?: { adapter_version: string; read_only: boolean; [key: string]: unknown };
    };
  };
  review: { title: string; consequence: string; fields: Field[] };
  adapter: string;
  credentials: Field[];
  status: string;
  approval?: { digest: string; approved_at: string; expires_at: string };
  receipt?: Receipt;
  can_approve: boolean;
  can_execute: boolean;
  can_reject: boolean;
  server_time: string;
};
export type InboxItem = {
  id: string; title: string; adapter: string; action: string; resource: string;
  created_at: string; status: string; problem?: string;
};
export type Inbox = { schema_version: 1; items: InboxItem[]; total: number; truncated: boolean };
export type RunResult = { receipt: Receipt; response?: unknown };
export type Reply<T> = { data: T | null; error: string | null };
export interface Client {
  inbox(): Promise<Reply<Inbox>>;
  review(id: string): Promise<Reply<Review>>;
  approve(id: string, digest: string): Promise<Reply<unknown>>;
  reject(id: string): Promise<Reply<unknown>>;
  execute(id: string): Promise<Reply<RunResult>>;
}

// Make invisible direction/control characters visible in all approval text.
// React renders the returned string as text; request content is never HTML.
export function visible(value: string): string {
  return value.replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f\u200b-\u200f\u202a-\u202e\u2060-\u206f\ufeff]/g,
    c => `⟦U+${c.charCodeAt(0).toString(16).toUpperCase().padStart(4, "0")}⟧`);
}

export function formatted(value: unknown): string {
  return visible(JSON.stringify(value, null, 2) ?? "");
}
