from __future__ import annotations

from lxml import etree


class ValidationError(Exception):
    pass


def validate_finrep_fixture(xml_text: str) -> None:
    """Fixture-equivalent FINREP check (arelle optional later)."""
    root = etree.fromstring(xml_text.encode("utf-8"))
    ns = {"xbrl": "http://www.xbrl.org/2003/instance"}
    if root.tag != "{http://www.xbrl.org/2003/instance}xbrl" and not root.tag.endswith("xbrl"):
        # accept default NS xbrl
        if "xbrl" not in root.tag:
            raise ValidationError("not an xbrl root")
    if root.find(".//{http://www.xbrl.org/2003/instance}context") is None and root.find("context") is None:
        # try without ns
        contexts = root.xpath(".//*[local-name()='context']")
        if not contexts:
            raise ValidationError("missing context")
    facts = root.xpath(".//*[local-name()='mi_cash' or local-name()='mi_loans' or local-name()='mi_deposits']")
    if len(facts) < 3:
        raise ValidationError("missing FINREP F01 monetary facts")
    tv = root.xpath(".//*[local-name()='taxonomyVersion']")
    if not tv or not (tv[0].text or "").strip():
        raise ValidationError("missing taxonomyVersion pin")
    _ = ns


def validate_anacredit(xml_text: str) -> None:
    root = etree.fromstring(xml_text.encode("utf-8"))
    if not root.tag.endswith("AnaCreditTable2"):
        raise ValidationError("not AnaCreditTable2")
    if not root.get("taxonomyVersion"):
        raise ValidationError("missing taxonomyVersion")
    instruments = root.xpath(".//*[local-name()='Instrument']")
    if not instruments:
        raise ValidationError("no instruments")


def validate_dora(xml_text: str) -> None:
    root = etree.fromstring(xml_text.encode("utf-8"))
    if not root.tag.endswith("DORAICTRegister"):
        raise ValidationError("not DORAICTRegister")
    if not root.get("taxonomyVersion"):
        raise ValidationError("missing taxonomyVersion")
    assets = root.xpath(".//*[local-name()='ICTAsset']")
    if not assets:
        raise ValidationError("no ICT assets")


def validate_report(report_type: str, xml_text: str) -> None:
    if report_type == "FINREP_F01":
        validate_finrep_fixture(xml_text)
    elif report_type == "ANACREDIT_T2":
        validate_anacredit(xml_text)
    elif report_type == "DORA_ICT":
        validate_dora(xml_text)
    else:
        raise ValidationError(f"unknown report_type {report_type}")
