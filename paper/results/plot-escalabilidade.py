# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.8"]
# ///
"""Gera paper/figures/escalabilidade.pdf a partir dos dados de escalabilidade.

Fonte dos dados: o CSV bruto da varredura de 5 a 200 contêineres
(20261007-211639/scalability/scalability.csv, gerado por scripts/run-experiments.sh S),
que é a origem da Tabela tab:escalabilidade do artigo. Cada ponto é uma única
execução de 5000 requisições (concorrência 50).

Uso: uv run paper/results/plot-escalabilidade.py
"""

from __future__ import annotations

import csv
from pathlib import Path

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt

HERE = Path(__file__).resolve().parent
CSV = HERE / "20261007-211639" / "scalability" / "scalability.csv"
OUT = HERE.parent / "figures" / "escalabilidade.pdf"

XTICKS = [5, 50, 100, 150, 200]


def load_rows() -> list[tuple[int, float, float, float, float, int]]:
    rows = []
    with CSV.open(encoding="utf-8") as fh:
        for r in csv.DictReader(fh):
            rows.append(
                (
                    int(r["n_instances"]),
                    float(r["p50_ms"]),
                    float(r["p95_ms"]),
                    float(r["p99_ms"]),
                    float(r["err_pct"]),
                    int(r["mem_used_mb"]),
                )
            )
    if not rows:
        raise SystemExit(f"nenhuma linha de escalabilidade encontrada em {CSV}")
    return sorted(rows)


def main() -> None:
    rows = load_rows()
    n = [r[0] for r in rows]
    p50 = [r[1] for r in rows]
    p95 = [r[2] for r in rows]
    p99 = [r[3] for r in rows]
    mem = [r[5] for r in rows]

    plt.rcParams.update(
        {
            "font.size": 9,
            "axes.labelsize": 9,
            "legend.fontsize": 8,
            "xtick.labelsize": 8,
            "ytick.labelsize": 8,
            "axes.spines.top": False,
            "axes.spines.right": False,
            "pdf.fonttype": 42,
        }
    )

    fig, (ax_lat, ax_mem) = plt.subplots(1, 2, figsize=(6.3, 2.5))

    # Preto e branco: tons de cinza distintos + marcadores e traços diferentes.
    series = [
        ("p99", p99, "#222222", "-", "o"),
        ("p95", p95, "#555555", "--", "s"),
        ("p50", p50, "#888888", ":", "^"),
    ]
    for label, ys, color, ls, marker in series:
        ax_lat.plot(
            n,
            ys,
            color=color,
            linestyle=ls,
            marker=marker,
            markersize=4,
            linewidth=1.4,
            label=label,
        )
        ax_lat.annotate(
            label,
            (n[-1], ys[-1]),
            xytext=(4, 0),
            textcoords="offset points",
            va="center",
            fontsize=8,
        )
    ax_lat.set_xlabel("Contêineres em execução")
    ax_lat.set_ylabel("Latência de /meta-data (ms)")
    ax_lat.set_ylim(0, 140)
    ax_lat.set_xlim(0, 230)
    ax_lat.set_xticks(XTICKS)
    ax_lat.grid(axis="y", color="#dddddd", linewidth=0.6)
    ax_lat.legend(
        loc="lower right",
        bbox_to_anchor=(1.0, 0.22),
        frameon=False,
        ncol=3,
        columnspacing=1.0,
        handlelength=2.2,
    )
    ax_lat.set_title("(a) Latência", loc="left", fontsize=9)

    ax_mem.plot(
        n, mem, color="#222222", linestyle="-", marker="o", markersize=4, linewidth=1.4
    )
    ax_mem.set_xlabel("Contêineres em execução")
    ax_mem.set_ylabel("Memória utilizada no host (MB)")
    ax_mem.set_ylim(0, 6000)
    ax_mem.set_xlim(0, 230)
    ax_mem.set_xticks(XTICKS)
    ax_mem.grid(axis="y", color="#dddddd", linewidth=0.6)
    ax_mem.set_title("(b) Memória do host", loc="left", fontsize=9)

    fig.tight_layout(w_pad=2.0)
    OUT.parent.mkdir(parents=True, exist_ok=True)
    fig.savefig(OUT)
    print(f"gravado {OUT} ({len(rows)} pontos: n={n})")


if __name__ == "__main__":
    main()
