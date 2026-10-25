#!/usr/bin/env python3
"""Detect editorial source drift; optionally maintain one review issue per article.

No upstream code is executed, no article is rewritten and no baseline is advanced.
Default mode only prints JSON. Publishing is restricted to an explicit repository.
"""

import argparse
import fnmatch
import json
import os
from pathlib import Path
import re
import subprocess
import sys
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[1]
REPOSITORY = re.compile(r"[A-Za-z0-9][A-Za-z0-9-]{0,38}/[A-Za-z0-9][A-Za-z0-9_.-]{0,99}")
SHA = re.compile(r"[0-9a-f]{40}")
ISSUE_AUTHOR = "github-actions[bot]"


def github(method, endpoint, payload=None, select=None):
    command = ["gh", "api", "--method", method, endpoint]
    if payload is not None:
        command += ["--input", "-"]
    if select:
        command += ["--jq", select]
    try:
        result = subprocess.run(command, input=json.dumps(payload) if payload is not None else None,
                                capture_output=True, text=True, check=True, timeout=60)
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired) as error:
        # Never print API payloads, tokens or untrusted error bodies.
        raise RuntimeError(f"GitHub request failed: {method} {endpoint}") from error
    return json.loads(result.stdout)


def load_sources(path, root=ROOT):
    document = json.loads(Path(path).read_text(encoding="utf-8"))
    if not isinstance(document, dict) or set(document) != {"version", "sources"} or document["version"] != 1:
        raise ValueError("unsupported editorial source manifest")
    sources = document["sources"]
    if not isinstance(sources, list) or not sources:
        raise ValueError("sources must be a nonempty list")
    seen = set()
    required = {"slug", "article", "repository", "branch", "reviewed_commit", "watch_paths"}
    for source in sources:
        if not isinstance(source, dict) or set(source) != required:
            raise ValueError("invalid source fields")
        slug = source["slug"]
        if not isinstance(slug, str) or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", slug) or slug in seen:
            raise ValueError("invalid or duplicate source slug")
        seen.add(slug)
        if source["article"] != f"internal/resources/editorial/{slug}.md":
            raise ValueError("article path must match its editorial slug")
        if not isinstance(source["repository"], str) or not REPOSITORY.fullmatch(source["repository"]):
            raise ValueError("invalid source repository")
        if not isinstance(source["reviewed_commit"], str) or not SHA.fullmatch(source["reviewed_commit"]):
            raise ValueError("reviewed_commit must be a full lowercase commit SHA")
        branch = source["branch"]
        if not isinstance(branch, str) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_./-]*", branch):
            raise ValueError("invalid source branch")
        if any(part in ("", ".", "..") for part in branch.split("/")):
            raise ValueError("invalid source branch segments")
        patterns = source["watch_paths"]
        if not isinstance(patterns, list) or not patterns:
            raise ValueError("watch_paths must be a nonempty list")
        for pattern in patterns:
            if (not isinstance(pattern, str) or not pattern or pattern.startswith("/")
                    or ".." in pattern.split("/") or any(ord(char) < 32 for char in pattern)):
                raise ValueError("invalid watched path")
        pin = f"https://github.com/{source['repository']}/tree/{source['reviewed_commit']}"
        if pin not in (root / source["article"]).read_text(encoding="utf-8"):
            raise ValueError("article must link its reviewed source commit")
    return sources


def inspect(source, api=github):
    repository, base = source["repository"], source["reviewed_commit"]
    head = api("GET", f"repos/{repository}/commits/{quote(source['branch'], safe='')}",
               select="{sha: .sha}")["sha"]
    if not isinstance(head, str) or not SHA.fullmatch(head):
        raise ValueError("upstream did not return a full commit SHA")
    report = {"slug": source["slug"], "repository": repository, "reviewed_commit": base,
              "head_commit": head, "comparison_status": "identical", "changed_paths": [],
              "files_truncated": False, "needs_review": False}
    if head == base:
        return report
    comparison = api("GET", f"repos/{repository}/compare/{base}...{head}?per_page=100&page=1",
                     select="{status, files: [.files[] | {filename, previous_filename}]}")
    files = comparison["files"]
    if comparison["status"] not in ("ahead", "behind", "diverged", "identical"):
        raise ValueError("invalid upstream comparison status")
    if not isinstance(files, list):
        raise ValueError("upstream comparison has no file list")
    changed = set()
    for item in files:
        if not isinstance(item, dict) or not isinstance(item.get("filename"), str):
            raise ValueError("invalid upstream file record")
        for key in ("filename", "previous_filename"):
            name = item.get(key)
            if name is not None:
                if not isinstance(name, str):
                    raise ValueError("invalid upstream filename")
                if any(fnmatch.fnmatchcase(name, pattern) for pattern in source["watch_paths"]):
                    changed.add(name)
    report.update(comparison_status=comparison["status"], changed_paths=sorted(changed),
                  files_truncated=len(files) >= 300)
    # GitHub compare exposes at most 300 changed files (on the first page only).
    # An incomplete comparison or rewritten history must never look up-to-date.
    report["needs_review"] = bool(changed) or report["files_truncated"] or comparison["status"] != "ahead"
    return report


def issue_body(report):
    slug, repository = report["slug"], report["repository"]
    base, head = report["reviewed_commit"], report["head_commit"]
    # Do not allow filenames to inject Markdown fences, mentions or newlines.
    paths = [json.dumps(path, ensure_ascii=True).replace("`", "\\u0060").replace("@", "\\u0040")
             for path in report["changed_paths"][:40]]
    listing = "\n".join(f"- `{path}`" for path in paths) or "- No matched path visible; inspect history/limits."
    return f"""<!-- editorial-source:{slug} -->
## Upstream changed: editorial review needed

Article: `internal/resources/editorial/{slug}.md`
Reviewed commit: `{base}`
Observed upstream: `{head}`
[Compare source changes](https://github.com/{repository}/compare/{base}...{head})

Comparison status: `{report['comparison_status']}`. File list may be incomplete: `{report['files_truncated']}`.
Relevant paths (up to 40 shown):
{listing}

This is a drift notification, not proof that any claim is incorrect or that hardware behavior is validated.

- [ ] Review source contracts, article claims, diagrams and tests against the comparison.
- [ ] Distinguish builds/native tests from physical hardware evidence; do not invent validation.
- [ ] Update the Markdown article and `updated_at` only for a real editorial update; preserve `published_at`.
- [ ] Deliberately advance `reviewed_commit` in `doc/editorial_sources.json` and the article's source link after review.
- [ ] Run the site checks and merge through a reviewed PR, then close this issue.

The watcher never rewrites prose, advances the baseline, publishes content or closes this issue automatically.
This body is managed by the watcher; keep manual review notes in comments.
"""


def publish(report, repository, api=github):
    if not REPOSITORY.fullmatch(repository):
        raise ValueError("invalid issue repository")
    if not report["needs_review"]:
        return
    marker = f"<!-- editorial-source:{report['slug']} -->"
    title = f"Editorial review: {report['slug']}"
    body = issue_body(report)
    for page in range(1, 101):
        issues = api("GET", f"repos/{repository}/issues?state=open&per_page=100&page={page}",
                     select="[.[] | {number, body, pull_request, title, author: .user.login}]")
        for issue in issues:
            if (not issue.get("pull_request") and issue.get("author") == ISSUE_AUTHOR
                    and issue.get("title") == title and marker in (issue.get("body") or "")):
                if issue["body"] != body:
                    api("PATCH", f"repos/{repository}/issues/{int(issue['number'])}", {"body": body})
                return
        if len(issues) < 100:
            api("POST", f"repos/{repository}/issues",
                {"title": title, "body": body})
            return
    raise RuntimeError("issue pagination limit reached; refusing to create a possible duplicate")


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", type=Path, default=ROOT / "doc/editorial_sources.json")
    parser.add_argument("--publish", action="store_true", help="create/update review issues")
    parser.add_argument("--issue-repo", help="explicit owner/repo receiving issues")
    args = parser.parse_args(argv)
    if args.publish and not args.issue_repo:
        parser.error("--publish requires --issue-repo")
    if args.publish and (os.environ.get("GITHUB_ACTIONS") != "true"
                         or os.environ.get("GITHUB_REPOSITORY") != args.issue_repo):
        parser.error("--publish is reserved for this repository's Actions workflow; dispatch it instead")
    try:
        for source in load_sources(args.manifest):
            report = inspect(source)
            print(json.dumps(report, sort_keys=True))
            if args.publish:
                publish(report, args.issue_repo)
    except (OSError, ValueError, KeyError, TypeError, RuntimeError) as error:
        print(f"editorial source check failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
