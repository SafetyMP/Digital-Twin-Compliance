// INT-R003: Sensitive twin data requires Analyst or Reporter role.
@id("INT-R003")
@description("Sensitive twin data access requires role=Analyst or role=Reporter")
permit(
  principal,
  action == DigitalTwin::Action::"view",
  resource
)
when {
  resource is DigitalTwin::TwinData &&
  (
    resource.sensitivity != "high" ||
    principal in DigitalTwin::Role::"Analyst" ||
    principal in DigitalTwin::Role::"Reporter"
  )
};
