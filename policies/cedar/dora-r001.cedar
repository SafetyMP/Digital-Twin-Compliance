// DORA-R001: Critical ICT contract changes require ICTRiskManager role.
@id("DORA-R001")
@description("Critical ICT contract change requires role=ICTRiskManager")
permit(
  principal,
  action == DigitalTwin::Action::"modify",
  resource
)
when {
  resource is DigitalTwin::ICTContract &&
  (
    resource.criticality != "critical" ||
    principal in DigitalTwin::Role::"ICTRiskManager"
  )
};
