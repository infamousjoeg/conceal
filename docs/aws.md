# Let an agent list backup names, without giving it AWS keys

[Back to Conceal](../README.md) · [Desktop approval inbox](desktop.md) · [Adapter contribution guide](adapters.md)

Suppose you want an assistant to check which monthly reports were backed up to S3. It needs a list of names under `reports/2026/`; it does not need the ability to download, overwrite, or delete files.

Conceal's first AWS adapter prepares **one S3 ListObjectsV2 request** for an allowed bucket and prefix. You review it, approve once, and run it. Conceal signs the request after approval and returns the service response. AWS CLI and Summon are not required.

This adapter currently supports general-purpose S3 buckets in **US East (N. Virginia), `us-east-1`; US West (Oregon), `us-west-2`; and Europe (Ireland), `eu-west-1`**. It does not support other regions, AWS China/GovCloud, directory buckets, access points, custom endpoints, or S3-compatible services. Use the bucket's actual region: redirects are denied.

## Practice first

Open [desktop practice mode](desktop.md) to try a complete synthetic listing. No AWS account or credentials are needed.

For CLI preparation, these examples also work without credentials. They validate local policy only; they do not check whether the example bucket exists or is accessible:

```sh
conceal action validate examples/s3-contract.json
conceal action prepare --contract examples/s3-contract.json --operation aws.s3.objects.list --resource example-backups --input examples/list-s3-objects.json
```

Review the result in the desktop inbox, then **reject it** to complete the exercise without making a request.

## Configure a real listing

Copy [the sample contract](../examples/s3-contract.json), replace the bucket name and region, and choose a prefix:

```json
{
  "schema_version": 2,
  "name": "backup-inventory",
  "actions": ["aws.s3.objects.list"],
  "approval_ttl_seconds": 300,
  "adapter": "aws",
  "credentials": {
    "access_key_id": "aws/access-key-id",
    "secret_access_key": "aws/secret-access-key",
    "session_token": "aws/session-token"
  },
  "scope": {
    "region": "us-east-1",
    "buckets": [{ "name": "YOUR-BUCKET", "prefix": "reports/" }]
  }
}
```

The prefix must be present. An empty string explicitly permits listing the whole bucket. Prefixes use S3 string-prefix semantics: `reports/` includes keys beginning with those exact characters, while `reports` would also include `reports-old/`. This contract permits one chosen bucket per request.

Use existing, appropriately scoped **temporary AWS credentials** from your trusted authentication workflow. Store all three values through Conceal's hidden prompts:

```sh
conceal set aws/access-key-id
conceal set aws/secret-access-key
conceal set aws/session-token
```

The session token is required for temporary credentials. If an existing workflow uses long-lived credentials, remove `session_token` from the contract; do not store an empty placeholder. Conceal does not sign in to AWS, discover profiles, invoke AWS CLI, assume roles, renew sessions, or copy credentials from your environment. Use your established AWS CLI/IAM Identity Center workflow for those tasks. An expired AWS session remains expired even if the Conceal approval is valid.

The AWS principal needs **`s3:ListBucket`** for the chosen bucket. Restrict its IAM policy to the required bucket and, where appropriate, an `s3:prefix` condition. Conceal's local contract complements service-side IAM; it does not reduce the underlying key's permissions outside Conceal.

Prepare with your real contract and bucket:

```sh
conceal action prepare --contract my-s3-contract.json --operation aws.s3.objects.list --resource YOUR-BUCKET --input examples/list-s3-objects.json
```

Then refresh the desktop, review **bucket, region, prefix, maximum results, and credential names**, approve once, and run. The [CLI review/approve/execute commands](actions.md#prepare-review-approve-execute) work too.

## What comes back

The limit is 1–100 keys, default 30. There is one page, with no automatic pagination or retries. The response is S3's XML, returned as a JSON string by the CLI and displayed as text by the desktop. Object names are URL-encoded by S3 because the request uses `encoding-type=url`.

Conceal does not download object contents. Names, sizes, timestamps, and other returned metadata can still be private. Inspect the response before sharing it with an agent. A successful HTTP response is recorded; Conceal does not parse XML to certify its contents.

AWS request charges may apply. Conceal is local and has no service fee; AWS pricing and account/free-tier eligibility still apply. Review [S3 pricing](https://aws.amazon.com/s3/pricing/) for your bucket's region.

For API permissions, response details, and service limits, see AWS's [ListObjectsV2 documentation](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html).
