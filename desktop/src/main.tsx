import { useCallback, useEffect, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import { liveClient, native } from "./client";
import { PracticeClient } from "./practice";
import { formatted, visible, type Client, type Inbox, type Review, type RunResult } from "./model";
import "./style.css";

const labels: Record<string, string> = {
  pending: "Needs review", approved: "Approved once", expired: "Approval expired", stale: "Request expired",
  consumed: "Approval used", rejected: "Rejected", invalid: "Cannot review", busy: "In progress",
};
const time = (value: string) => new Date(value).toLocaleString();

function App() {
  const [practice, setPractice] = useState(!native);
  const practiceClient = useRef(new PracticeClient());
  const client: Client = practice ? practiceClient.current : liveClient;
  const [inbox, setInbox] = useState<Inbox | null>(null);
  const [id, setID] = useState<string | null>(null);
  const [review, setReview] = useState<Review | null>(null);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [checked, setChecked] = useState(false);
  const [result, setResult] = useState<RunResult | null>(null);
  const [now, setNow] = useState(Date.now());
  const generation = useRef(0);
  const clockOffset = useRef(0);
  const heading = useRef<HTMLHeadingElement>(null);

  const load = useCallback(async (selected: string | null, focus = false) => {
    const generationID = ++generation.current;
    setLoading(true); setReview(null); setChecked(false); setError("");
    try {
      const list = await client.inbox();
      if (generation.current !== generationID) return;
      if (list.error || !list.data) throw new Error(list.error || "The inbox could not be loaded.");
      setInbox(list.data);
      const next = selected && list.data.items.some(item => item.id === selected) ? selected : list.data.items[0]?.id ?? null;
      setID(next);
      if (next) {
        const response = await client.review(next);
        if (generation.current !== generationID) return;
        if (response.error || !response.data) throw new Error(response.error || "The request could not be reviewed.");
        if (response.data.schema_version !== 1 || response.data.proposal.request.id !== next) throw new Error("Unsupported or mismatched review document.");
        clockOffset.current = Date.parse(response.data.server_time) - Date.now();
        setNow(Date.now());
        setReview(response.data);
        if (focus) requestAnimationFrame(() => heading.current?.focus());
      }
    } catch (e) {
      if (generation.current === generationID) setError(visible(String(e)));
    } finally {
      if (generation.current === generationID) setLoading(false);
    }
  }, [client]);

  useEffect(() => { setInbox(null); setNotice(""); setResult(null); void load(null); return () => { generation.current++; }; }, [load]);
  useEffect(() => { const timer = setInterval(() => setNow(Date.now()), 1000); return () => clearInterval(timer); }, []);

  const select = (selected: string) => { setNotice(""); setResult(null); void load(selected, true); };
  const changeMode = () => {
    if (busy) return;
    setID(null); setReview(null); setChecked(false); setPractice(value => !value);
  };
  const act = async (verb: "approve" | "reject" | "execute") => {
    if (!review || busy || loading) return;
    // Capture exactly the displayed snapshot. Never substitute a freshly fetched
    // digest at approval time: the engine rejects any intervening change.
    const displayed = review.proposal;
    setBusy(true); setError(""); setNotice(""); setChecked(false); setResult(null);
    let failure = "";
    try {
      const response = verb === "approve" ? await client.approve(displayed.request.id, displayed.digest) :
        verb === "reject" ? await client.reject(displayed.request.id) : await client.execute(displayed.request.id);
      failure = response.error || "";
      if (verb === "execute" && response.data) setResult(response.data as RunResult);
      if (!failure) setNotice(verb === "approve" ? "Approved once. Nothing has run yet." :
        verb === "reject" ? "Rejected. This request cannot run." :
          practice ? "Practice complete. No service was contacted." : "Action completed. This approval is now used.");
    } catch (e) { failure = String(e); }
    await load(displayed.request.id);
    if (failure) setError(visible(failure));
    setBusy(false);
  };
  const expired = review?.approval ? now + clockOffset.current >= Date.parse(review.approval.expires_at) : false;
  const status = expired && review?.status === "approved" ? "expired" : review?.status ?? "";
  const request = review?.proposal.request;

  return <div className="app">
    <header className="topbar">
      <a className="wordmark" href="#" onClick={e => { e.preventDefault(); }}>conceal<span className="wordmark-dot" aria-hidden="true">.</span></a>
      <span className="app-name">Approval inbox</span>
      <div className="mode-controls">
        <span className="mode">{practice ? "Practice" : "This computer"}</span>
        {native && <button className="text-button" disabled={busy} onClick={changeMode}>{practice ? "Open real inbox" : "Try practice mode"}</button>}
      </div>
    </header>
    {practice && <div className="practice-banner"><strong>Practice mode</strong><span>Synthetic requests. No credentials, files, or services are accessed.</span>
      <button className="text-button" disabled={busy} onClick={() => { practiceClient.current.reset(); setNotice(""); setResult(null); void load(null); }}>Reset practice</button></div>}
    <div className="workspace">
      <aside className="inbox-pane" aria-label="Requests">
        <div className="pane-heading"><h1>Requests</h1><button className="text-button" disabled={busy || loading} onClick={() => { setResult(null); setNotice(""); void load(id); }}>Refresh</button></div>
        <p className="inbox-hint">Review what will happen before a credential is used.</p>
        <nav aria-label="Choose a request" className="request-list">
          {inbox?.items.map(item => <button key={item.id} className={`request-item ${id === item.id ? "selected" : ""}`}
            aria-current={id === item.id ? "true" : undefined} disabled={busy} onClick={() => select(item.id)}>
            <span className="request-service">{item.adapter === "aws" ? "Amazon S3" : item.adapter === "github" ? "GitHub" : visible(item.adapter || "Unknown adapter")}</span>
            <strong>{visible(item.title)}</strong><span className="request-resource">{visible(item.resource || item.id)}</span>
            <span className={`status status-${item.status}`}>{labels[item.status] || visible(item.status)}</span>
          </button>)}
        </nav>
        {inbox?.truncated && <p className="small">Showing the latest {inbox.items.length} of {inbox.total} requests. Use the CLI to inspect an older request by ID.</p>}
        <div className="trust-note"><strong>Your decision, one request.</strong><p>An approval expires and can be used once. Changes require another review.</p></div>
      </aside>
      <main className="review-pane" aria-busy={loading || busy}>
        {error && <div className="message error" role="alert"><strong>Could not continue</strong><p>{error}</p><p>Refresh to review the current state. If dispatch started, check the remote service before preparing another request.</p></div>}
        {notice && <div className="message success" role="status">{notice}</div>}
        {loading && <p className="loading" role="status">Checking the request and its contract…</p>}
        {!loading && !review && !error && <section className="empty">
          <span className="eyebrow">Start with a request</span><h2>Your inbox is ready.</h2>
          <p>Prepare an action with Conceal. It will appear here for you to review, approve, or reject.</p>
          <pre>conceal action prepare --contract contract.json{"\n"}  --operation github.issues.list --resource OWNER/REPO</pre>
          <p className="small">Use <code>conceal action example --adapter github</code> to create a contract, then customize it. Refresh this inbox after preparing a request.</p>
          {!practice && <button className="primary" onClick={changeMode}>Try a practice request</button>}
        </section>}
        {review && request && <article className="review-document">
          <div className="document-top"><span className="eyebrow">{review.adapter === "aws" ? "Amazon S3" : review.adapter === "github" ? "GitHub" : visible(review.adapter)}</span>
            <span className={`status status-${status}`}>{labels[status] || visible(status)}</span></div>
          <h2 ref={heading} tabIndex={-1}>{visible(review.review.title)}</h2>
          <p className="consequence">{visible(review.review.consequence)}</p>
          <dl className="review-fields">{review.review.fields.map((field, index) => <div className="review-field" key={index}>
            <dt>{visible(field.label)}</dt><dd>{visible(field.value) || <span className="muted">(empty)</span>}</dd>
          </div>)}</dl>
          <section className="credential-section">
            <h3>Credentials used only when this runs</h3><p className="small">These are names in your operating system credential store. Values stay out of this review.</p>
            <dl>{review.credentials.map((field, index) => <div className="credential" key={index}><dt>{visible(field.label)}</dt><dd><code>{visible(field.value)}</code></dd></div>)}</dl>
          </section>
          <section className="approval-terms"><h3>One exact request. One use.</h3>
            <p>Your approval is valid for {request.contract.approval_ttl_seconds} seconds. It covers the action, input, resource, contract, and prepared HTTP request.</p>
            <p className="small">{request.contract.schema_version === 1 ? "Legacy contract: use schema 2 for approvals that also bind the adapter version and structured review." : "This schema-2 request also binds the adapter version and complete review."}</p>
            <p className="small">No automatic redirects or retries. Credential values may be rotated under the same names. This app does not verify who authored the request.</p>
          </section>
          <details className="technical"><summary>Inspect the complete request</summary>
            <dl><div><dt>Prepared</dt><dd>{time(request.created_at)}</dd></div>
              <div><dt>Contract file</dt><dd><code>{visible(request.contract_path)}</code></dd></div>
              <div><dt>Request ID</dt><dd><code>{request.id}</code></dd></div>
              <div><dt>Reviewed digest</dt><dd><code data-testid="digest">{review.proposal.digest}</code></dd></div></dl>
            <h4>HTTP request before authentication</h4><pre>{formatted(request.http)}</pre>
            <h4>Full proposal</h4><pre>{formatted(review.proposal)}</pre>
          </details>
          {review.receipt && <section className="receipt"><h3>{review.receipt.outcome === "revoked" ? "Request rejected" : review.receipt.outcome === "succeeded" ? "Action succeeded" : "Approval consumed"}</h3>
            <p>{review.receipt.outcome === "revoked" ? "No request was dispatched." : review.receipt.outcome === "succeeded" ?
              "The service returned success. This approval cannot run again." : "The outcome may be uncertain. Check the remote service before creating another request."}</p>
            <p className="small">{time(review.receipt.time)}{review.receipt.http_status ? ` · HTTP ${review.receipt.http_status}` : ""}</p>
            <details><summary>View receipt</summary><pre>{formatted(review.receipt)}</pre></details>
          </section>}
          {result?.response !== undefined && <details className="response"><summary>View {practice ? "simulated " : ""}service response</summary>
            <p className="small">Service output is untrusted data. It may contain private information.</p><pre>{formatted(result.response)}</pre></details>}
          <div className="decision-bar">
            {status === "pending" && <label className="confirmation"><input type="checkbox" checked={checked} disabled={busy || loading} onChange={e => setChecked(e.target.checked)} />
              <span>I reviewed the action, target, and all fields above.</span></label>}
            {status === "approved" && review.approval && <p className="approval-clock">Approved until {time(review.approval.expires_at)}. Nothing runs until you choose Run.</p>}
            {(status === "expired" || status === "stale") && <p>This request needs a fresh proposal and review. It cannot be renewed here.</p>}
            {status === "busy" && <p>This request is in progress or an earlier operation was interrupted. Refresh to check its state.</p>}
            {status === "consumed" && !review.receipt && <p>Approval was consumed, but there is no receipt. Check the remote service before preparing another request.</p>}
            <div className="decision-buttons">
              {review.can_reject && <button className="reject" disabled={busy || loading} onClick={() => void act("reject")}>Reject request</button>}
              {review.can_approve && <button className="primary" disabled={!checked || busy || loading} onClick={() => void act("approve")}>{busy ? "Saving…" : "Approve once"}</button>}
              {review.can_execute && <button className="primary" disabled={expired || busy || loading} onClick={() => void act("execute")}>{busy ? "Running…" : practice ? "Run practice action" : "Run approved action"}</button>}
            </div>
          </div>
        </article>}
        <footer className="operator-note">Keep this app and its approval files in your trusted operator account. Unrestricted software running as you can bypass a same-account approval workflow.</footer>
      </main>
    </div>
  </div>;
}

createRoot(document.getElementById("root")!).render(<App />);
