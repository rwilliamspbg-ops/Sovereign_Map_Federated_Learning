#!/usr/bin/env python3

import argparse
import json
import subprocess
from dataclasses import dataclass, field
from datetime import date
from pathlib import Path

COMMIT_WEIGHT = 5.0
ADDITION_WEIGHT = 0.05
DELETION_WEIGHT = 0.02
FILE_WEIGHT = 0.5


@dataclass
class ContributorStats:
    name: str
    email: str
    commits: int = 0
    additions: int = 0
    deletions: int = 0
    files: set[str] = field(default_factory=set)
    points: int = 0
    rank: int = 0
    tier: str = ""

    def calculate_points(self) -> int:
        score = (
            (self.commits * COMMIT_WEIGHT)
            + (self.additions * ADDITION_WEIGHT)
            + (self.deletions * DELETION_WEIGHT)
            + (len(self.files) * FILE_WEIGHT)
        )
        self.points = int(round(score))
        return self.points


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate contributor rankings and reward points from git history"
    )
    parser.add_argument("--repo", default=".", help="Path to git repository")
    parser.add_argument("--branch", default="HEAD", help="Branch or revision range root")
    parser.add_argument("--since", default="30 days ago", help="Git --since value")
    parser.add_argument("--until", default="now", help="Git --until value")
    parser.add_argument("--top", type=int, default=25, help="Top N contributors to include")
    parser.add_argument(
        "--output-json",
        default="test-results/contributor-rankings/contributor_rankings.json",
        help="Output JSON path",
    )
    parser.add_argument(
        "--output-md",
        default="test-results/contributor-rankings/CONTRIBUTOR_RANKINGS.md",
        help="Output Markdown path",
    )
    return parser.parse_args()


def get_git_log(repo_path: Path, branch: str, since: str, until: str) -> str:
    cmd = [
        "git",
        "-C",
        str(repo_path),
        "log",
        branch,
        f"--since={since}",
        f"--until={until}",
        "--pretty=format:@@@%H|%an|%ae",
        "--numstat",
    ]
    result = subprocess.run(cmd, check=True, capture_output=True, text=True)
    return result.stdout


def tier_for_points(points: int) -> str:
    if points >= 500:
        return "Platinum"
    if points >= 250:
        return "Gold"
    if points >= 120:
        return "Silver"
    return "Bronze"


def build_rankings(log_output: str, top_n: int) -> list[ContributorStats]:
    contributors: dict[str, ContributorStats] = {}
    current_email = ""

    for raw_line in log_output.splitlines():
        line = raw_line.strip()
        if not line:
            continue

        if line.startswith("@@@"):
            _, author_name, author_email = line[3:].split("|", 2)
            current_email = author_email.lower()
            stats = contributors.get(current_email)
            if not stats:
                stats = ContributorStats(name=author_name, email=author_email)
                contributors[current_email] = stats
            stats.commits += 1
            continue

        parts = line.split("\t")
        if len(parts) != 3 or not current_email:
            continue

        added_raw, deleted_raw, file_path = parts
        stats = contributors[current_email]
        stats.files.add(file_path)

        if added_raw.isdigit():
            stats.additions += int(added_raw)
        if deleted_raw.isdigit():
            stats.deletions += int(deleted_raw)

    ranking = list(contributors.values())
    for stats in ranking:
        stats.calculate_points()

    ranking.sort(
        key=lambda entry: (
            entry.points,
            entry.commits,
            entry.additions,
            -entry.deletions,
        ),
        reverse=True,
    )

    for index, stats in enumerate(ranking, start=1):
        stats.rank = index
        stats.tier = tier_for_points(stats.points)

    return ranking[:top_n]


def write_json(path: Path, rankings: list[ContributorStats], since: str, until: str) -> None:
    payload = {
        "generated_on": date.today().isoformat(),
        "window": {"since": since, "until": until},
        "weights": {
            "commit": COMMIT_WEIGHT,
            "addition": ADDITION_WEIGHT,
            "deletion": DELETION_WEIGHT,
            "file": FILE_WEIGHT,
        },
        "rankings": [
            {
                "rank": entry.rank,
                "name": entry.name,
                "email": entry.email,
                "tier": entry.tier,
                "points": entry.points,
                "commits": entry.commits,
                "additions": entry.additions,
                "deletions": entry.deletions,
                "files_touched": len(entry.files),
            }
            for entry in rankings
        ],
    }
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2), encoding="utf-8")


def write_markdown(path: Path, rankings: list[ContributorStats], since: str, until: str) -> None:
    lines = [
        "# Contributor Rankings & Reward Points",
        "",
        f"- Window: `{since}` → `{until}`",
        "- Points formula: `commits*5 + additions*0.05 + deletions*0.02 + files_touched*0.5`",
        "",
        "| Rank | Contributor | Tier | Points | Commits | Additions | Deletions | Files |",
        "|------|-------------|------|--------|---------|-----------|-----------|-------|",
    ]

    for entry in rankings:
        lines.append(
            f"| {entry.rank} | {entry.name} ({entry.email}) | {entry.tier} | {entry.points} | "
            f"{entry.commits} | {entry.additions} | {entry.deletions} | {len(entry.files)} |"
        )

    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> int:
    args = parse_args()
    repo_path = Path(args.repo).resolve()
    log_output = get_git_log(repo_path, args.branch, args.since, args.until)
    rankings = build_rankings(log_output, args.top)

    json_path = (repo_path / args.output_json).resolve()
    md_path = (repo_path / args.output_md).resolve()

    write_json(json_path, rankings, args.since, args.until)
    write_markdown(md_path, rankings, args.since, args.until)

    print(f"Generated contributor rankings: {md_path}")
    print(f"Generated contributor points JSON: {json_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())