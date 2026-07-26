// Fixture-generated proposal stub — NOT auto-deployed (CR-3 / ADR-013 D38)
permit (
  principal in Role::"Analyst",
  action == Action::"viewAlert",
  resource
) when {
  resource.tenantId == "00000000-0000-0000-0000-000000000001"
};
