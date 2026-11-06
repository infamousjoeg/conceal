#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::Serialize;
use serde_json::Value;
use tauri_plugin_shell::ShellExt;

#[derive(Serialize)]
struct Reply {
    data: Option<Value>,
    error: Option<String>,
}

// Only these fixed action verbs cross the webview boundary. No shell command,
// executable path, environment, or state-directory override comes from the UI.
fn arguments(verb: &str, id: Option<&str>, digest: Option<&str>) -> Result<Vec<String>, String> {
    let mut args = vec!["action".to_owned()];
    match verb {
        "inbox" if id.is_none() && digest.is_none() => {
            args.extend(["inbox".to_owned(), "--limit".to_owned(), "100".to_owned()]);
        }
        "review" | "revoke" | "execute" | "approve" => {
            let id = id.ok_or("A request ID is required")?;
            if !hexadecimal(id, 32) {
                return Err("Invalid request ID".to_owned());
            }
            args.extend([verb.to_owned(), id.to_owned()]);
            if verb == "approve" {
                let digest = digest.ok_or("The reviewed digest is required")?;
                if !hexadecimal(digest, 64) {
                    return Err("Invalid reviewed digest".to_owned());
                }
                args.extend(["--digest".to_owned(), digest.to_owned()]);
            } else if digest.is_some() {
                return Err("Unexpected digest".to_owned());
            }
        }
        _ => return Err("Unsupported action".to_owned()),
    }
    Ok(args)
}

fn hexadecimal(value: &str, length: usize) -> bool {
    value.len() == length
        && value
            .bytes()
            .all(|c| c.is_ascii_digit() || (b'a'..=b'f').contains(&c))
}

async fn call(app: tauri::AppHandle, args: Vec<String>) -> Result<Reply, String> {
    let mut command = app
        .shell()
        .sidecar("conceal-engine")
        .map_err(|_| "The bundled Conceal engine is unavailable")?;
    // A trusted launcher may select an isolated test/operator profile. Agent
    // proposals and web content cannot choose or change this directory.
    let mut args = args;
    if let Some(directory) = std::env::var_os("CONCEAL_DESKTOP_STATE_DIR") {
        let directory = directory
            .into_string()
            .map_err(|_| "Invalid operator state directory")?;
        args.extend(["--state-dir".to_owned(), directory]);
    }
    command = command.args(args);
    let output = command
        .output()
        .await
        .map_err(|_| "Could not start the bundled Conceal engine")?;
    if output.stdout.len() > 32 * 1024 * 1024 || output.stderr.len() > 1024 * 1024 {
        return Err("Engine output exceeded the review limit; refresh the inbox".to_owned());
    }
    let data = if output.stdout.is_empty() {
        None
    } else {
        Some(
            serde_json::from_slice(&output.stdout)
                .map_err(|_| "The engine returned an invalid document")?,
        )
    };
    let error = if output.status.success() {
        None
    } else {
        let text = String::from_utf8_lossy(&output.stderr);
        Some(if text.trim().is_empty() {
            "The action could not complete. Refresh its review before continuing.".to_owned()
        } else {
            text.trim().to_owned()
        })
    };
    Ok(Reply { data, error })
}

#[tauri::command]
async fn inbox(app: tauri::AppHandle) -> Result<Reply, String> {
    call(app, arguments("inbox", None, None)?).await
}

#[tauri::command]
async fn review(app: tauri::AppHandle, id: String) -> Result<Reply, String> {
    call(app, arguments("review", Some(&id), None)?).await
}

#[tauri::command]
async fn approve(app: tauri::AppHandle, id: String, digest: String) -> Result<Reply, String> {
    call(app, arguments("approve", Some(&id), Some(&digest))?).await
}

#[tauri::command]
async fn reject(app: tauri::AppHandle, id: String) -> Result<Reply, String> {
    call(app, arguments("revoke", Some(&id), None)?).await
}

#[tauri::command]
async fn execute(app: tauri::AppHandle, id: String) -> Result<Reply, String> {
    call(app, arguments("execute", Some(&id), None)?).await
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            inbox, review, approve, reject, execute
        ])
        .run(tauri::generate_context!())
        .expect("could not run Conceal");
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn approval_carries_only_the_displayed_digest() {
        let args = arguments("approve", Some(&"a".repeat(32)), Some(&"b".repeat(64))).unwrap();
        assert_eq!(
            args,
            [
                "action",
                "approve",
                &"a".repeat(32),
                "--digest",
                &"b".repeat(64)
            ]
        );
    }

    #[test]
    fn rejects_commands_paths_and_malformed_identifiers() {
        for id in [
            "../request",
            "--help",
            "abc;whoami",
            "",
            "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        ] {
            assert!(arguments("execute", Some(id), None).is_err());
        }
        assert!(arguments("run", None, None).is_err());
        assert!(arguments("approve", Some(&"a".repeat(32)), None).is_err());
        assert!(arguments("approve", Some(&"a".repeat(32)), Some(&"g".repeat(64))).is_err());
        assert!(arguments("review", Some(&"a".repeat(32)), Some(&"b".repeat(64))).is_err());
    }
}
