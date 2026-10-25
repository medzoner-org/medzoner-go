import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

SCRIPT = Path(__file__).resolve().parents[1] / "editorial_source_watch.py"
SPEC = importlib.util.spec_from_file_location("editorial_source_watch", SCRIPT)
watch = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(watch)
BASE, HEAD = "1" * 40, "2" * 40
SOURCE = {
    "slug": "camera", "article": "internal/resources/editorial/camera.md",
    "repository": "Example/camera", "branch": "main", "reviewed_commit": BASE,
    "watch_paths": ["src/**", "web/**", "version.txt"],
}


def inspect(files=(), status="ahead", head=HEAD):
    def api(method, endpoint, payload=None, select=None):
        if "/commits/" in endpoint:
            return {"sha": head}
        return {"status": status, "files": list(files)}
    return watch.inspect(SOURCE, api)


class SourceWatchTests(unittest.TestCase):
    def test_repository_manifest_and_article_pin_are_valid(self):
        sources = watch.load_sources(watch.ROOT / "doc/editorial_sources.json")
        self.assertEqual(len(sources), 1)
        self.assertEqual(sources[0]["slug"], "resilient-multi-client-esp32-camera")

    def test_unchanged_head_does_not_fetch_comparison(self):
        calls = []
        def api(*args, **kwargs):
            calls.append(args)
            return {"sha": BASE}
        self.assertFalse(watch.inspect(SOURCE, api)["needs_review"])
        self.assertEqual(len(calls), 1)

    def test_tracked_changes_and_renames_require_review(self):
        for file in ({"filename": "src/main.cpp"},
                     {"filename": "archive/main.cpp", "previous_filename": "src/main.cpp"},
                     {"filename": "web/index.js"}, {"filename": "version.txt"}):
            with self.subTest(file=file):
                self.assertTrue(inspect([file])["needs_review"])

    def test_unrelated_changes_are_not_claimed_as_editorial_drift(self):
        report = inspect([{"filename": "LICENSE"}])
        self.assertFalse(report["needs_review"])
        self.assertEqual(report["reviewed_commit"], BASE)

    def test_github_file_cap_never_silently_hides_drift(self):
        report = inspect([{"filename": f"other/{index}"} for index in range(300)])
        self.assertTrue(report["files_truncated"])
        self.assertTrue(report["needs_review"])

    def test_rewritten_or_behind_history_requires_review(self):
        for status in ("behind", "diverged", "identical"):
            self.assertTrue(inspect(status=status)["needs_review"])

    def test_invalid_comparison_and_api_failure_are_not_clean_results(self):
        with self.assertRaises(ValueError):
            inspect(status="unknown")
        with self.assertRaises(ValueError):
            inspect([{}])
        with self.assertRaises(ValueError):
            inspect(head="main")
        def api(*args, **kwargs):
            raise RuntimeError("API unavailable")
        with self.assertRaises(RuntimeError):
            watch.inspect(SOURCE, api)

    def test_manifest_rejects_invalid_paths_duplicate_slugs_and_unpinned_article(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            article = root / SOURCE["article"]
            article.parent.mkdir(parents=True)
            article.write_text(f"https://github.com/Example/camera/tree/{BASE}")
            manifest = root / "sources.json"
            valid = {"version": 1, "sources": [SOURCE]}
            manifest.write_text(json.dumps(valid))
            self.assertEqual(watch.load_sources(manifest, root)[0]["slug"], "camera")
            for key, value in (("article", "../secrets"), ("reviewed_commit", "main"),
                               ("repository", "../camera"), ("branch", "main/../other"),
                               ("watch_paths", []), ("watch_paths", ["../src/**"])):
                data = copy.deepcopy(valid)
                data["sources"][0][key] = value
                manifest.write_text(json.dumps(data))
                with self.subTest(key=key, value=value), self.assertRaises(ValueError):
                    watch.load_sources(manifest, root)
            manifest.write_text(json.dumps({"version": 1, "sources": [SOURCE, SOURCE]}))
            with self.assertRaises(ValueError):
                watch.load_sources(manifest, root)
            manifest.write_text(json.dumps(valid))
            article.write_text("Article without reviewed source link")
            with self.assertRaises(ValueError):
                watch.load_sources(manifest, root)

    def test_publish_creates_one_issue_when_none_exists(self):
        calls = []
        def api(method, endpoint, payload=None, select=None):
            calls.append((method, endpoint, payload))
            return [] if method == "GET" else {"number": 42}
        watch.publish(inspect([{"filename": "src/main.cpp"}]), "Example/site", api)
        self.assertEqual([call[0] for call in calls], ["GET", "POST"])
        self.assertIn("<!-- editorial-source:camera -->", calls[-1][2]["body"])
        self.assertIn(f"/{BASE}...{HEAD}", calls[-1][2]["body"])

    def test_publish_updates_existing_issue_and_unchanged_body_is_noop(self):
        report = inspect([{"filename": "src/main.cpp"}])
        body = watch.issue_body(report)
        for existing_body, expected in (("<!-- editorial-source:camera --> old", ["GET", "PATCH"]),
                                         (body, ["GET"])):
            calls = []
            def api(method, endpoint, payload=None, select=None):
                calls.append(method)
                return [{"number": 42, "body": existing_body, "author": watch.ISSUE_AUTHOR,
                         "title": "Editorial review: camera"}] if method == "GET" else {}
            watch.publish(report, "Example/site", api)
            self.assertEqual(calls, expected)

    def test_publish_paginates_and_does_not_treat_pr_as_review_issue(self):
        calls = []
        def api(method, endpoint, payload=None, select=None):
            calls.append((method, endpoint))
            if endpoint.endswith("page=1"):
                return [{"number": n, "body": "<!-- editorial-source:camera -->", "pull_request": {"url": "pr"}}
                        for n in range(100)]
            return [{"number": 999, "body": "<!-- editorial-source:camera --> old",
                     "author": watch.ISSUE_AUTHOR, "title": "Editorial review: camera"}] if method == "GET" else {}
        watch.publish(inspect([{"filename": "src/main.cpp"}]), "Example/site", api)
        self.assertEqual([call[0] for call in calls], ["GET", "GET", "PATCH"])

    def test_clean_report_has_no_issue_side_effects(self):
        def api(*args, **kwargs):
            self.fail("clean report must not access issues")
        watch.publish(inspect(head=BASE), "Example/site", api)

    def test_untrusted_filename_cannot_inject_issue_markdown_or_mentions(self):
        body = watch.issue_body(inspect([{"filename": "src/```\n@everyone"}]))
        self.assertNotIn("```", body)
        self.assertNotIn("@everyone", body)
        self.assertIn("\\u0060", body)

    def test_default_cli_only_prints_and_does_not_publish(self):
        with patch.object(watch, "load_sources", return_value=[SOURCE]), \
                patch.object(watch, "inspect", return_value=inspect([{"filename": "src/main.cpp"}])), \
                patch.object(watch, "publish") as publish, patch("builtins.print"):
            self.assertEqual(watch.main([]), 0)
            publish.assert_not_called()

    def test_forged_marker_or_unrelated_title_is_never_overwritten(self):
        for author, title in (("someone-else", "Editorial review: camera"),
                              (watch.ISSUE_AUTHOR, "Unrelated issue")):
            calls = []
            def api(method, endpoint, payload=None, select=None):
                calls.append(method)
                return [{"number": 7, "author": author, "title": title,
                         "body": "<!-- editorial-source:camera --> copied"}] if method == "GET" else {}
            watch.publish(inspect([{"filename": "src/main.cpp"}]), "Example/site", api)
            self.assertEqual(calls, ["GET", "POST"])

    def test_publication_is_not_available_with_an_arbitrary_local_token(self):
        with patch.dict(watch.os.environ, {}, clear=True), \
                patch.object(watch, "load_sources") as sources, \
                patch.object(watch, "inspect") as inspect_source, \
                patch.object(watch, "publish") as publish, patch("sys.stderr"):
            with self.assertRaises(SystemExit):
                watch.main(["--publish", "--issue-repo", "Example/site"])
            sources.assert_not_called()
            inspect_source.assert_not_called()
            publish.assert_not_called()


if __name__ == "__main__":
    unittest.main()
