from __future__ import annotations

from datetime import date
from xml.sax.saxutils import escape


def generate_finrep_f01(taxonomy_version: str, maps: dict[str, str], figures: dict[str, float]) -> str:
    cash = figures.get("assets.cash", 1_000_000.0)
    loans = figures.get("assets.loans", 5_000_000.0)
    deposits = figures.get("liab.deposits", 4_500_000.0)
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<xbrl xmlns="http://www.xbrl.org/2003/instance"
      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
      xmlns:cect="http://digitaltwin.local/finrep/f01/{escape(taxonomy_version)}"
      xsi:schemaLocation="http://digitaltwin.local/finrep/f01/{escape(taxonomy_version)} finrep-f01-fixture.xsd">
  <context id="c1">
    <entity><identifier scheme="http://digitaltwin.local/lei">CECTSEED000000001</identifier></entity>
    <period><instant>{date.today().isoformat()}</instant></period>
  </context>
  <unit id="uEUR"><measure>iso4217:EUR</measure></unit>
  <cect:{maps.get("assets.cash", "mi_cash")} contextRef="c1" unitRef="uEUR" decimals="0">{cash:.0f}</cect:{maps.get("assets.cash", "mi_cash")}>
  <cect:{maps.get("assets.loans", "mi_loans")} contextRef="c1" unitRef="uEUR" decimals="0">{loans:.0f}</cect:{maps.get("assets.loans", "mi_loans")}>
  <cect:{maps.get("liab.deposits", "mi_deposits")} contextRef="c1" unitRef="uEUR" decimals="0">{deposits:.0f}</cect:{maps.get("liab.deposits", "mi_deposits")}>
  <cect:taxonomyVersion contextRef="c1">{escape(taxonomy_version)}</cect:taxonomyVersion>
</xbrl>
"""


def generate_anacredit_t2(taxonomy_version: str, maps: dict[str, str], instruments: list[dict]) -> str:
    rows = []
    for inst in instruments:
        itype = inst.get("type", "loan")
        code = maps.get(itype, maps.get("loan", "CRDT"))
        rows.append(
            f'    <Instrument id="{escape(str(inst.get("id", "i1")))}" '
            f'type="{escape(code)}" '
            f'notional="{float(inst.get("notional", 0)):.2f}" '
            f'currency="{escape(str(inst.get("currency", "EUR")))}"/>'
        )
    body = "\n".join(rows) if rows else '    <Instrument id="seed-1" type="CRDT" notional="1000000.00" currency="EUR"/>'
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<AnaCreditTable2 xmlns="http://digitaltwin.local/anacredit/t2"
                 taxonomyVersion="{escape(taxonomy_version)}"
                 asOf="{date.today().isoformat()}">
{body}
</AnaCreditTable2>
"""


def generate_dora_ict(taxonomy_version: str, maps: dict[str, str], assets: list[dict]) -> str:
    rows = []
    for asset in assets:
        name = asset.get("name", "unknown")
        code = maps.get(name, maps.get("core-banking", "ICT-CORE"))
        rows.append(
            f'    <ICTAsset name="{escape(str(name))}" '
            f'function="{escape(code)}" '
            f'criticality="{escape(str(asset.get("criticality", "high")))}"/>'
        )
    body = "\n".join(rows)
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<DORAICTRegister xmlns="http://digitaltwin.local/dora/ict"
                 taxonomyVersion="{escape(taxonomy_version)}"
                 asOf="{date.today().isoformat()}">
{body}
</DORAICTRegister>
"""
