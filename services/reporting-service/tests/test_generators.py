from reporting_service.generators import generate_anacredit_t2, generate_dora_ict, generate_finrep_f01
from reporting_service.validate import validate_report


def test_finrep_validates():
    maps = {"assets.cash": "mi_cash", "assets.loans": "mi_loans", "liab.deposits": "mi_deposits"}
    xml = generate_finrep_f01("2024.1", maps, {"assets.cash": 1, "assets.loans": 2, "liab.deposits": 3})
    validate_report("FINREP_F01", xml)


def test_anacredit_validates():
    xml = generate_anacredit_t2("2024.1", {"loan": "CRDT"}, [{"id": "a", "type": "loan", "notional": 10}])
    validate_report("ANACREDIT_T2", xml)


def test_dora_validates():
    xml = generate_dora_ict("2024.1", {"kafka": "ICT-MSG"}, [{"name": "kafka"}])
    validate_report("DORA_ICT", xml)
