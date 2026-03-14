#!/usr/bin/env python3
"""Plot CDF of transaction processing times from latency_*.txt files.

Usage:
    python3 plot_cdf.py latency_node1.txt latency_node2.txt ... --label "3 nodes 0.5Hz" --out cdf.png
    python3 plot_cdf.py scenario1/*.txt scenario2/*.txt --labels "S1" "S2" --out cdf.png
"""

import argparse
import numpy as np
import matplotlib.pyplot as plt
import os


def load_latencies(paths):
    vals = []
    for p in paths:
        with open(p) as f:
            for line in f:
                line = line.strip()
                if line:
                    vals.append(float(line))
    return np.array(vals)


def plot_cdf(ax, latencies, label):
    sorted_ms = np.sort(latencies)
    n = len(sorted_ms)
    # percentile rank of each point (1..99)
    pcts = np.linspace(1, 99, n)
    # clip to 1–99th percentile range
    lo, hi = np.percentile(sorted_ms, 1), np.percentile(sorted_ms, 99)
    mask = (sorted_ms >= lo) & (sorted_ms <= hi)
    ax.plot(sorted_ms[mask], pcts[mask], label=label)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("files", nargs="+", help="latency_*.txt files (all combined per scenario)")
    parser.add_argument("--label", default=None, help="Single label for all files combined")
    parser.add_argument("--out", default="cdf.png", help="Output image path")
    args = parser.parse_args()

    fig, ax = plt.subplots(figsize=(8, 5))

    latencies = load_latencies(args.files)
    label = args.label or ", ".join(os.path.basename(f) for f in args.files)
    plot_cdf(ax, latencies, label)

    ax.set_xlabel("Transaction Processing Time (ms)")
    ax.set_ylabel("Percentile")
    ax.set_ylim(1, 99)
    ax.set_title("CDF of Transaction Processing Time")
    ax.legend()
    ax.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(args.out, dpi=150)
    print(f"Saved {args.out}")


if __name__ == "__main__":
    main()
