import { invoke, isTauri } from "@tauri-apps/api/core";
import type { Client, Inbox, Reply, Review, RunResult } from "./model";

export const native = isTauri();
export const liveClient: Client = {
  inbox: () => invoke<Reply<Inbox>>("inbox"),
  review: id => invoke<Reply<Review>>("review", { id }),
  approve: (id, digest) => invoke<Reply<unknown>>("approve", { id, digest }),
  reject: id => invoke<Reply<unknown>>("reject", { id }),
  execute: id => invoke<Reply<RunResult>>("execute", { id }),
};
