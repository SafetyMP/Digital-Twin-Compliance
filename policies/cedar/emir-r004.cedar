// EMIR-R004: Trade reporting requires TradeReporter role.
@id("EMIR-R004")
@description("Trade reporting data access requires role=TradeReporter")
permit(
  principal,
  action == DigitalTwin::Action::"report",
  resource
)
when {
  resource is DigitalTwin::TradeReport &&
  principal in DigitalTwin::Role::"TradeReporter"
};
